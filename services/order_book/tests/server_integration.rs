use tonic::Request;

use order_book::server::OrderBookServer;
use order_book::server::order_book_service::orderbook_service_server::OrderbookService;
use order_book::server::order_book_service::{
    AddOrderRequest, CancelOrderRequest, InjectPriceRequest,
};

fn limit_order(id: &str, symbol: &str, side: i32, price: i64, qty: i64) -> AddOrderRequest {
    AddOrderRequest {
        id: id.to_string(),
        user_id: format!("user-{}", id),
        symbol: symbol.to_string(),
        side,
        order_type: 1,
        time_in_force: 1,
        qty,
        price,
        stop_price: 0,
    }
}

async fn tick(server: &OrderBookServer, symbol: &str, price_cents: i64) -> i64 {
    server
        .inject_price(Request::new(InjectPriceRequest {
            symbol: symbol.to_string(),
            price_cents,
        }))
        .await
        .unwrap()
        .into_inner()
        .fills
}

const BUY: i32 = 0;
const SELL: i32 = 1;

#[tokio::test]
async fn add_order_rests_and_reports_accepted() {
    let server = OrderBookServer::new(None);
    tick(&server, "AAPL", 11_000).await;

    let resp = server
        .add_order(Request::new(limit_order("o1", "AAPL", BUY, 10_000, 5)))
        .await
        .unwrap()
        .into_inner();

    assert!(resp.accepted);
    assert!(resp.trades.is_empty());
}

#[tokio::test]
async fn crossing_user_orders_do_not_trade_with_each_other() {
    let server = OrderBookServer::new(None);
    tick(&server, "AAPL", 10_000).await;

    server
        .add_order(Request::new(limit_order("b1", "AAPL", BUY, 9_900, 5)))
        .await
        .unwrap();

    let resp = server
        .add_order(Request::new(limit_order("s1", "AAPL", SELL, 9_900, 5)))
        .await
        .unwrap()
        .into_inner();

    assert!(resp.accepted);
    assert_eq!(resp.trades.len(), 1);
    let trade = &resp.trades[0];
    assert_eq!(trade.ask_trade.as_ref().unwrap().order_id, "s1");
    assert_eq!(trade.bid_trade.as_ref().unwrap().order_id, "market");
    assert_eq!(trade.ask_trade.as_ref().unwrap().price, 10_000);

    let cancel = server
        .cancel_order(Request::new(CancelOrderRequest {
            order_id: "b1".to_string(),
            symbol: "AAPL".to_string(),
        }))
        .await
        .unwrap()
        .into_inner();
    assert!(cancel.cancelled);
}

#[tokio::test]
async fn symbols_are_isolated_from_each_other() {
    let server = OrderBookServer::new(None);
    tick(&server, "AAPL", 10_000).await;

    server
        .add_order(Request::new(limit_order("b1", "AAPL", BUY, 9_000, 5)))
        .await
        .unwrap();

    let resp = server
        .add_order(Request::new(limit_order("s1", "TSLA", SELL, 9_000, 5)))
        .await
        .unwrap()
        .into_inner();
    assert!(resp.trades.is_empty());

    assert_eq!(tick(&server, "TSLA", 9_500).await, 1);
    assert_eq!(tick(&server, "AAPL", 10_001).await, 0);
}

#[tokio::test]
async fn market_order_fills_at_injected_price() {
    let server = OrderBookServer::new(None);
    tick(&server, "AAPL", 10_000).await;

    let mut req = limit_order("m1", "AAPL", BUY, 0, 5);
    req.order_type = 0;
    let resp = server
        .add_order(Request::new(req))
        .await
        .unwrap()
        .into_inner();

    assert_eq!(resp.trades.len(), 1);
    let trade = &resp.trades[0];
    assert_eq!(trade.bid_trade.as_ref().unwrap().order_id, "m1");
    assert_eq!(trade.bid_trade.as_ref().unwrap().price, 10_000);
    assert_eq!(trade.ask_trade.as_ref().unwrap().order_id, "market");
}

#[tokio::test]
async fn resting_limit_fills_on_a_crossing_tick() {
    let server = OrderBookServer::new(None);
    tick(&server, "AAPL", 10_000).await;

    server
        .add_order(Request::new(limit_order("b1", "AAPL", BUY, 9_500, 5)))
        .await
        .unwrap();

    assert_eq!(tick(&server, "AAPL", 9_600).await, 0, "not crossed yet");
    assert_eq!(tick(&server, "AAPL", 9_500).await, 1, "crossed: fills");
}

#[tokio::test]
async fn ioc_limit_order_kills_when_not_marketable() {
    let server = OrderBookServer::new(None);
    tick(&server, "AAPL", 10_000).await;

    let mut req = limit_order("i1", "AAPL", BUY, 9_000, 10);
    req.time_in_force = 2;
    let resp = server
        .add_order(Request::new(req))
        .await
        .unwrap()
        .into_inner();

    assert!(resp.accepted);
    assert!(resp.trades.is_empty());

    let cancel = server
        .cancel_order(Request::new(CancelOrderRequest {
            order_id: "i1".to_string(),
            symbol: "AAPL".to_string(),
        }))
        .await
        .unwrap()
        .into_inner();
    assert!(!cancel.cancelled);
}

#[tokio::test]
async fn cancel_resting_order_succeeds_once() {
    let server = OrderBookServer::new(None);
    tick(&server, "AAPL", 11_000).await;

    server
        .add_order(Request::new(limit_order("o1", "AAPL", BUY, 10_000, 5)))
        .await
        .unwrap();

    let cancel = |id: &str| {
        let req = CancelOrderRequest {
            order_id: id.to_string(),
            symbol: "AAPL".to_string(),
        };
        server.cancel_order(Request::new(req))
    };

    let first = cancel("o1").await.unwrap().into_inner();
    assert!(first.cancelled);

    let second = cancel("o1").await.unwrap().into_inner();
    assert!(!second.cancelled);
}

#[tokio::test]
async fn cancel_on_unknown_symbol_returns_not_cancelled() {
    let server = OrderBookServer::new(None);

    let resp = server
        .cancel_order(Request::new(CancelOrderRequest {
            order_id: "o1".to_string(),
            symbol: "NOPE".to_string(),
        }))
        .await
        .unwrap()
        .into_inner();

    assert!(!resp.cancelled);
}

#[tokio::test]
async fn concurrent_orders_on_one_symbol_serialize_correctly() {
    use std::sync::Arc;

    let server = Arc::new(OrderBookServer::new(None));
    tick(&server, "AAPL", 10_000).await;

    let mut handles = Vec::new();
    for i in 0..20 {
        let s = server.clone();
        handles.push(tokio::spawn(async move {
            s.add_order(Request::new(limit_order(
                &format!("b{}", i),
                "AAPL",
                BUY,
                10_000,
                1,
            )))
            .await
            .unwrap()
            .into_inner()
        }));
        let s = server.clone();
        handles.push(tokio::spawn(async move {
            s.add_order(Request::new(limit_order(
                &format!("s{}", i),
                "AAPL",
                SELL,
                10_000,
                1,
            )))
            .await
            .unwrap()
            .into_inner()
        }));
    }

    let mut total_traded = 0i64;
    for h in handles {
        let resp = h.await.unwrap();
        assert!(resp.accepted);
        total_traded += resp
            .trades
            .iter()
            .map(|t| t.bid_trade.as_ref().unwrap().quantity)
            .sum::<i64>();
    }

    assert_eq!(total_traded, 40, "all 40 marketable orders fill fully");
}

#[tokio::test]
async fn stop_limit_order_triggers_on_a_crossing_tick() {
    let server = OrderBookServer::new(None);
    tick(&server, "AAPL", 10_000).await;

    let mut stop = limit_order("st1", "AAPL", SELL, 9_400, 5);
    stop.order_type = 3;
    stop.stop_price = 9_500;
    let resp = server
        .add_order(Request::new(stop))
        .await
        .unwrap()
        .into_inner();
    assert!(resp.accepted);
    assert!(resp.trades.is_empty(), "stop must park, not trade");

    assert_eq!(tick(&server, "AAPL", 9_450).await, 1);
}

#[tokio::test]
async fn fok_order_kills_when_not_marketable_through_grpc() {
    let server = OrderBookServer::new(None);
    tick(&server, "AAPL", 10_000).await;

    let mut fok = limit_order("f1", "AAPL", BUY, 9_000, 10);
    fok.time_in_force = 3;
    let resp = server
        .add_order(Request::new(fok))
        .await
        .unwrap()
        .into_inner();

    assert!(resp.accepted);
    assert!(resp.trades.is_empty(), "FOK below the market must kill");
}

#[tokio::test]
async fn inject_price_rejects_non_positive_prices() {
    let server = OrderBookServer::new(None);
    let err = server
        .inject_price(Request::new(InjectPriceRequest {
            symbol: "AAPL".to_string(),
            price_cents: 0,
        }))
        .await
        .unwrap_err();
    assert_eq!(err.code(), tonic::Code::InvalidArgument);
}

mod worker_thread {
    use order_book::server::commands::{Command, OrderData};
    use order_book::server::workers::spawn_worker_for_symbol;
    use order_book::{OrderType, Side};
    use tokio::sync::oneshot;

    fn order_data(id: &str, side: Side, price: i64, qty: i64) -> OrderData {
        OrderData {
            id: id.to_string(),
            user_id: "u1".to_string(),
            side,
            price,
            quantity: qty,
            order_type: OrderType::GoodTillCancel,
            stop_price: None,
        }
    }

    fn send_inject(tx: &crossbeam_channel::Sender<Command>, price: i64) -> usize {
        let (resp_tx, resp_rx) = oneshot::channel();
        tx.send(Command::InjectPrice {
            price,
            resp: resp_tx,
        })
        .unwrap();
        resp_rx.blocking_recv().unwrap().trades.len()
    }

    #[test]
    fn worker_processes_add_and_cancel_commands() {
        let tx = spawn_worker_for_symbol("TEST".to_string());
        send_inject(&tx, 200);

        let (resp_tx, resp_rx) = oneshot::channel();
        tx.send(Command::AddOrder {
            order: order_data("o1", Side::Buy, 100, 5),
            resp: resp_tx,
        })
        .unwrap();
        let result = resp_rx.blocking_recv().unwrap();
        assert!(result.accepted);
        assert!(result.trades.is_empty());

        let (resp_tx, resp_rx) = oneshot::channel();
        tx.send(Command::CancelOrder {
            order_id: "o1".to_string(),
            resp: resp_tx,
        })
        .unwrap();
        assert!(resp_rx.blocking_recv().unwrap().cancelled);
    }

    #[test]
    fn worker_fills_resting_order_on_price_tick() {
        let tx = spawn_worker_for_symbol("TEST2".to_string());
        send_inject(&tx, 200);

        let (resp_tx, resp_rx) = oneshot::channel();
        tx.send(Command::AddOrder {
            order: order_data("b1", Side::Buy, 100, 5),
            resp: resp_tx,
        })
        .unwrap();
        assert!(resp_rx.blocking_recv().unwrap().trades.is_empty());

        assert_eq!(send_inject(&tx, 100), 1);
    }
}
