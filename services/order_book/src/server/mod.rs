pub mod commands;
pub mod workers;

use std::collections::HashMap;
use std::sync::Mutex;

use crossbeam_channel::Sender;
use tokio::sync::oneshot;
use tonic::{Request, Response, Status};

use order_book_service::orderbook_service_server::OrderbookService;
use order_book_service::{
    AddOrderRequest, AddOrderResponse, CancelOrderRequest, CancelOrderResponse, InjectPriceRequest,
    InjectPriceResponse, Trade as ProtoTrade, TradeInfo as ProtoTradeInfo,
};

use crate::server::commands::{Command, OrderData};
use crate::server::workers::spawn_worker_for_symbol;
use crate::{OrderType, Side};

pub mod order_book_service {
    tonic::include_proto!("order_book_service");
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
}
