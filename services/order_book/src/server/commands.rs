use tokio::sync::oneshot;

use crate::{DoneInfo, OrderType, Price, Quantity, Side, orderbook::trade::Trade};

#[derive(Debug)]
pub enum Command {
    AddOrder {
        order: OrderData,
        resp: oneshot::Sender<AddOrderResult>,
    },
}

#[derive(Debug, Clone)]
pub struct OrderData {
    pub id: String,
    pub user_id: String,
    pub side: Side,
    pub price: Price,
    pub quantity: Quantity,
    pub order_type: OrderType,
    pub stop_price: Option<Price>,
}

#[derive(Debug)]
pub struct AddOrderResult {
    pub trades: Vec<Trade>,
    pub accepted: bool,
    pub done: Vec<DoneInfo>,
}
