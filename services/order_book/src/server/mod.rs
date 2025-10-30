pub mod commands;
pub mod workers;

use std::collections::HashMap;
use std::sync::Mutex;

use chrono::{DateTime, Utc};
use crossbeam_channel::Sender;
use serde::Serialize;
use tokio::sync::oneshot;
use tonic::{Request, Response, Status};

use order_book_service::orderbook_service_server::OrderbookService;
use order_book_service::{
    AddOrderRequest, AddOrderResponse, CancelOrderRequest, CancelOrderResponse, InjectPriceRequest,
    InjectPriceResponse, Trade as ProtoTrade, TradeInfo as ProtoTradeInfo,
};

use crate::orderbook::trade::Trade;
use crate::server::commands::{Command, OrderData};
use crate::server::workers::spawn_worker_for_symbol;
use crate::{DoneInfo, OrderType, Side};

pub mod order_book_service {
    tonic::include_proto!("order_book_service");
}

const FILL_ROUTING_KEY: &str = "order.fill";
const DONE_ROUTING_KEY: &str = "order.done";

#[derive(Debug, Serialize, PartialEq)]
pub struct FillEvent {
    pub trade_id: String,
    pub symbol: String,
    pub order_id: String,
    pub counter_order_id: String,
    pub user_id: String,
    pub side: i16,
    pub qty: i64,
    pub price_cents: i64,
    pub liquidity: String,
    pub executed_at: DateTime<Utc>,
}

#[derive(Debug, Serialize, PartialEq)]
pub struct DoneEvent {
    pub order_id: String,
    pub user_id: String,
    pub reason: String,
    pub unfilled_qty: i64,
}

pub fn build_fill_events(
    symbol: &str,
    trade: &Trade,
    executed_at: DateTime<Utc>,
) -> Vec<FillEvent> {
    let liquidity = |order_id: &str| {
        if order_id == trade.taker_order_id {
            "taker".to_string()
        } else {
            "maker".to_string()
        }
    };

    let mut events = Vec::with_capacity(2);
    if trade.bid_trade.order_id != crate::orderbook::book::MARKET_COUNTERPARTY {
        events.push(FillEvent {
            trade_id: trade.trade_id.clone(),
            symbol: symbol.to_string(),
            order_id: trade.bid_trade.order_id.clone(),
            counter_order_id: trade.ask_trade.order_id.clone(),
            user_id: trade.bid_trade.user_id.clone(),
            side: 0,
            qty: trade.bid_trade.quantity,
            price_cents: trade.bid_trade.price,
            liquidity: liquidity(&trade.bid_trade.order_id),
            executed_at,
        });
    }
    if trade.ask_trade.order_id != crate::orderbook::book::MARKET_COUNTERPARTY {
        events.push(FillEvent {
            trade_id: trade.trade_id.clone(),
            symbol: symbol.to_string(),
            order_id: trade.ask_trade.order_id.clone(),
            counter_order_id: trade.bid_trade.order_id.clone(),
            user_id: trade.ask_trade.user_id.clone(),
            side: 1,
            qty: trade.ask_trade.quantity,
            price_cents: trade.ask_trade.price,
            liquidity: liquidity(&trade.ask_trade.order_id),
            executed_at,
        });
    }
    events
}

pub fn build_done_event(done: &DoneInfo) -> DoneEvent {
    DoneEvent {
        order_id: done.order_id.clone(),
        user_id: done.user_id.clone(),
        reason: done.reason.as_str().to_string(),
        unfilled_qty: done.unfilled_qty,
    }
}

pub struct OrderBookServer {
    workers: Mutex<HashMap<String, Sender<Command>>>,
    rmq_channel: Option<lapin::Channel>,
}

impl OrderBookServer {
    pub fn new(rmq_channel: Option<lapin::Channel>) -> Self {
        Self {
            workers: Mutex::new(HashMap::new()),
            rmq_channel,
        }
    }

    fn get_or_spawn_worker(&self, symbol: &str) -> Sender<Command> {
        let mut workers = self.workers.lock().unwrap();
        workers
            .entry(symbol.to_string())
            .or_insert_with(|| spawn_worker_for_symbol(symbol.to_string()))
            .clone()
    }

    async fn publish_events(&self, symbol: &str, trades: &[Trade], done: &[DoneInfo]) {
        let Some(channel) = &self.rmq_channel else {
            return;
        };

        let now = Utc::now();
        for trade in trades {
            for event in build_fill_events(symbol, trade, now) {
                self.publish_json(channel, FILL_ROUTING_KEY, &event).await;
            }
        }
        for info in done {
            self.publish_json(channel, DONE_ROUTING_KEY, &build_done_event(info))
                .await;
        }
    }

    async fn publish_json<T: Serialize>(
        &self,
        channel: &lapin::Channel,
        routing_key: &str,
        event: &T,
    ) {
        let payload = match serde_json::to_vec(event) {
            Ok(p) => p,
            Err(e) => {
                eprintln!("Failed to serialize event: {}", e);
                return;
            }
        };

        let result = channel
            .basic_publish(
                "order.events",
                routing_key,
                lapin::options::BasicPublishOptions::default(),
                &payload,
                lapin::BasicProperties::default()
                    .with_delivery_mode(2)
                    .with_content_type("application/json".into()),
            )
            .await;

        match result {
            Ok(confirm) => {
                if let Err(e) = confirm.await {
                    eprintln!("RabbitMQ publish confirm error: {}", e);
                }
            }
            Err(e) => {
                eprintln!("Failed to publish event to RabbitMQ: {}", e);
            }
        }
    }
}

fn proto_order_type_to_internal(ot: i32, tif: i32) -> OrderType {
    match ot {
        0 => OrderType::Market,
        _ => match tif {
            2 => OrderType::ImmediateOrCancel,
            3 => OrderType::FillOrKill,
            _ => OrderType::GoodTillCancel,
        },
    }
}

fn proto_side_to_internal(side: i32) -> Side {
    match side {
        1 => Side::Sell,
        _ => Side::Buy,
    }
}

fn proto_stop_price(ot: i32, stop_price: i64) -> Option<i64> {
    match ot {
        2 | 3 if stop_price > 0 => Some(stop_price),
        _ => None,
    }
}

#[tonic::async_trait]
impl OrderbookService for OrderBookServer {
    async fn add_order(
        &self,
        request: Request<AddOrderRequest>,
    ) -> Result<Response<AddOrderResponse>, Status> {
        let req = request.into_inner();
        let symbol = req.symbol.clone();

        let tx = self.get_or_spawn_worker(&symbol);

        let order_data = OrderData {
            id: req.id,
            user_id: req.user_id,
            side: proto_side_to_internal(req.side),
            price: req.price,
            quantity: req.qty,
            order_type: proto_order_type_to_internal(req.order_type, req.time_in_force),
            stop_price: proto_stop_price(req.order_type, req.stop_price),
        };

        let (resp_tx, resp_rx) = oneshot::channel();

        let send_result = tokio::task::spawn_blocking(move || {
            tx.send(Command::AddOrder {
                order: order_data,
                resp: resp_tx,
            })
        })
        .await;

        if send_result.is_err() || send_result.unwrap().is_err() {
            return Err(Status::internal("failed to dispatch order to worker"));
        }

        let result = resp_rx
            .await
            .map_err(|_| Status::internal("worker did not respond"))?;

        self.publish_events(&symbol, &result.trades, &result.done)
            .await;

        let proto_trades: Vec<ProtoTrade> = result
            .trades
            .iter()
            .map(|t| ProtoTrade {
                bid_trade: Some(ProtoTradeInfo {
                    order_id: t.bid_trade.order_id.clone(),
                    price: t.bid_trade.price,
                    quantity: t.bid_trade.quantity,
                }),
                ask_trade: Some(ProtoTradeInfo {
                    order_id: t.ask_trade.order_id.clone(),
                    price: t.ask_trade.price,
                    quantity: t.ask_trade.quantity,
                }),
            })
            .collect();

        Ok(Response::new(AddOrderResponse {
            accepted: result.accepted,
            trades: proto_trades,
        }))
    }

    async fn cancel_order(
        &self,
        _request: Request<CancelOrderRequest>,
    ) -> Result<Response<CancelOrderResponse>, Status> {
        Err(Status::unimplemented("cancel_order"))
    }

    async fn inject_price(
        &self,
        _request: Request<InjectPriceRequest>,
    ) -> Result<Response<InjectPriceResponse>, Status> {
        Err(Status::unimplemented("inject_price"))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::orderbook::types::{DoneReason, TradeInfo};

    #[test]
    fn market_proto_type_maps_to_market() {
        assert!(matches!(
            proto_order_type_to_internal(0, 0),
            OrderType::Market
        ));
        assert!(matches!(
            proto_order_type_to_internal(0, 2),
            OrderType::Market
        ));
    }

    #[test]
    fn limit_proto_type_maps_by_time_in_force() {
        assert!(matches!(
            proto_order_type_to_internal(1, 0),
            OrderType::GoodTillCancel
        ));
        assert!(matches!(
            proto_order_type_to_internal(1, 1),
            OrderType::GoodTillCancel
        ));
        assert!(matches!(
            proto_order_type_to_internal(1, 2),
            OrderType::ImmediateOrCancel
        ));
        assert!(matches!(
            proto_order_type_to_internal(1, 3),
            OrderType::FillOrKill
        ));
    }

    #[test]
    fn stop_proto_types_carry_their_trigger() {
        assert_eq!(proto_stop_price(2, 9_500), Some(9_500));
        assert_eq!(proto_stop_price(3, 9_500), Some(9_500));
        assert_eq!(proto_stop_price(1, 9_500), None);
        assert_eq!(proto_stop_price(2, 0), None);
    }

    #[test]
    fn proto_side_maps_buy_and_sell() {
        assert!(matches!(proto_side_to_internal(0), Side::Buy));
        assert!(matches!(proto_side_to_internal(1), Side::Sell));
        assert!(matches!(proto_side_to_internal(42), Side::Buy));
    }

    fn sample_trade() -> Trade {
        Trade::new(
            TradeInfo {
                order_id: "b1".into(),
                user_id: "user-b".into(),
                price: 10_000,
                quantity: 5,
            },
            TradeInfo {
                order_id: "s1".into(),
                user_id: "user-s".into(),
                price: 10_000,
                quantity: 5,
            },
            "s1".into(),
        )
    }

    #[test]
    fn fill_events_carry_both_sides_with_liquidity_flags() {
        let trade = sample_trade();
        let now = Utc::now();
        let events = build_fill_events("AAPL", &trade, now);
        assert_eq!(events.len(), 2);
        let (bid, ask) = (&events[0], &events[1]);

        assert_eq!(bid.trade_id, trade.trade_id);
        assert_eq!(bid.symbol, "AAPL");
        assert_eq!(bid.order_id, "b1");
        assert_eq!(bid.counter_order_id, "s1");
        assert_eq!(bid.user_id, "user-b");
        assert_eq!(bid.side, 0);
        assert_eq!(bid.qty, 5);
        assert_eq!(bid.price_cents, 10_000);
        assert_eq!(bid.liquidity, "maker");

        assert_eq!(ask.order_id, "s1");
        assert_eq!(ask.counter_order_id, "b1");
        assert_eq!(ask.side, 1);
        assert_eq!(ask.liquidity, "taker");
        assert_eq!(ask.executed_at, now);
    }

    #[test]
    fn synthetic_market_side_emits_no_fill_event() {
        use crate::orderbook::book::MARKET_COUNTERPARTY;

        let trade = Trade::new(
            TradeInfo {
                order_id: "b1".into(),
                user_id: "user-b".into(),
                price: 10_000,
                quantity: 5,
            },
            TradeInfo {
                order_id: MARKET_COUNTERPARTY.into(),
                user_id: MARKET_COUNTERPARTY.into(),
                price: 10_000,
                quantity: 5,
            },
            "b1".into(),
        );

        let events = build_fill_events("AAPL", &trade, Utc::now());
        assert_eq!(events.len(), 1, "only the real order settles");
        assert_eq!(events[0].order_id, "b1");
        assert_eq!(events[0].side, 0);
        assert_eq!(events[0].counter_order_id, MARKET_COUNTERPARTY);
        assert_eq!(events[0].liquidity, "taker");
    }

    #[test]
    fn fill_event_serializes_to_the_consumer_contract() {
        let trade = sample_trade();
        let events = build_fill_events("AAPL", &trade, Utc::now());
        let bid = &events[0];

        let json: serde_json::Value =
            serde_json::from_slice(&serde_json::to_vec(&bid).unwrap()).unwrap();
        for key in [
            "trade_id",
            "symbol",
            "order_id",
            "counter_order_id",
            "user_id",
            "side",
            "qty",
            "price_cents",
            "liquidity",
            "executed_at",
        ] {
            assert!(json.get(key).is_some(), "missing key {key}");
        }
        assert_eq!(json["side"], 0);
        assert_eq!(json["qty"], 5);
        let ts = json["executed_at"].as_str().unwrap();
        assert!(ts.contains('T') && (ts.ends_with('Z') || ts.contains('+')));
    }

    #[test]
    fn done_event_serializes_reason_strings() {
        let done = DoneInfo {
            order_id: "o1".into(),
            user_id: "u1".into(),
            reason: DoneReason::IocExpired,
            unfilled_qty: 3,
        };
        let event = build_done_event(&done);
        let json: serde_json::Value =
            serde_json::from_slice(&serde_json::to_vec(&event).unwrap()).unwrap();
        assert_eq!(json["order_id"], "o1");
        assert_eq!(json["user_id"], "u1");
        assert_eq!(json["reason"], "ioc_expired");
        assert_eq!(json["unfilled_qty"], 3);
    }
}
