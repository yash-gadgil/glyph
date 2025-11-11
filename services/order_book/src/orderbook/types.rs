use std::{cell::RefCell, rc::Rc};

use crate::orderbook::{order::*, trade::Trade};

#[derive(Copy, Clone, Debug, PartialEq, Eq)]
pub enum OrderType {
    Market,
    GoodTillCancel,
    FillAndKill,
    ImmediateOrCancel,
    FillOrKill,
}

#[derive(Copy, Clone, Debug, PartialEq, Eq)]
pub enum Side {
    Buy,
    Sell,
}

pub type Price = i64;
pub type Quantity = i64;
pub type OrderId = String;
pub type UserId = String;

pub struct LevelInfo {
    pub price: Price,
    pub quantity: Quantity,
}

pub type LevelInfos = Vec<LevelInfo>;

pub type OrderPointer = Rc<RefCell<Order>>;
pub type OrderPointers = Vec<OrderPointer>;

#[derive(Debug, Clone)]
pub struct TradeInfo {
    pub order_id: OrderId,
    pub user_id: UserId,
    pub price: Price,
    pub quantity: Quantity,
}

pub type Trades = Vec<Trade>;

#[derive(Copy, Clone, Debug, PartialEq, Eq)]
pub enum DoneReason {
    Filled,
    Cancelled,
    IocExpired,
    FokKilled,
}

impl DoneReason {
    pub fn as_str(&self) -> &'static str {
        match self {
            DoneReason::Filled => "filled",
            DoneReason::Cancelled => "cancelled",
            DoneReason::IocExpired => "ioc_expired",
            DoneReason::FokKilled => "fok_killed",
        }
    }
}

#[derive(Debug, Clone)]
pub struct DoneInfo {
    pub order_id: OrderId,
    pub user_id: UserId,
    pub reason: DoneReason,
    pub unfilled_qty: Quantity,
}

#[derive(Debug, Default)]
pub struct MatchOutcome {
    pub trades: Trades,
    pub done: Vec<DoneInfo>,
    pub accepted: bool,
}

impl MatchOutcome {
    pub fn rejected() -> Self {
        Self {
            accepted: false,
            ..Default::default()
        }
    }

    pub fn merge(&mut self, other: MatchOutcome) {
        self.trades.extend(other.trades);
        self.done.extend(other.done);
    }
}

#[derive(Debug, Clone)]
pub struct StopOrder {
    pub order_id: OrderId,
    pub user_id: UserId,
    pub side: Side,
    pub trigger: Price,
    pub qty: Quantity,
    pub limit_price: Option<Price>,
}

#[derive(Debug)]
pub struct OrderEntry {
    pub order: OrderPointer,
}

pub struct OrderModify {
    pub order_id: OrderId,
    pub user_id: UserId,
    pub side: Side,
    pub price: Price,
    pub quantity: Quantity,
}

impl OrderModify {
    pub fn new(
        order_id: OrderId,
        user_id: UserId,
        side: Side,
        price: Price,
        quantity: Quantity,
    ) -> Self {
        Self {
            order_id,
            user_id,
            side,
            price,
            quantity,
        }
    }

    pub fn get_order_id(&self) -> &OrderId {
        &self.order_id
    }
    pub fn get_side(&self) -> &Side {
        &self.side
    }
    pub fn get_price(&self) -> Price {
        self.price
    }
    pub fn get_quantity(&self) -> Quantity {
        self.quantity
    }

    pub fn to_order_pointer(self, order_type: OrderType) -> OrderPointer {
        Rc::new(RefCell::new(Order::new(
            order_type,
            self.order_id,
            self.user_id,
            self.side,
            self.price,
            self.quantity,
        )))
    }
}

pub struct OrderbookLevelInfos {
    pub bids: LevelInfos,
    pub asks: LevelInfos,
}

impl OrderbookLevelInfos {
    pub fn new(bids: LevelInfos, asks: LevelInfos) -> Self {
        Self { bids, asks }
    }

    pub fn get_bids(&self) -> &LevelInfos {
        &self.bids
    }
    pub fn get_asks(&self) -> &LevelInfos {
        &self.asks
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn order_modify_exposes_its_fields() {
        let modify = OrderModify::new("o1".to_string(), "u1".to_string(), Side::Sell, 150, 25);
        assert_eq!(modify.get_order_id(), "o1");
        assert!(matches!(modify.get_side(), Side::Sell));
        assert_eq!(modify.get_price(), 150);
        assert_eq!(modify.get_quantity(), 25);
    }

    #[test]
    fn order_modify_converts_to_order_pointer() {
        let modify = OrderModify::new("o2".to_string(), "u2".to_string(), Side::Buy, 99, 7);
        let ptr = modify.to_order_pointer(OrderType::GoodTillCancel);
        let order = ptr.borrow();
        assert_eq!(order.get_order_id(), "o2");
        assert_eq!(order.get_user_id(), "u2");
        assert!(matches!(order.get_side(), Side::Buy));
        assert_eq!(order.get_price(), 99);
        assert_eq!(order.get_initial_quantity(), 7);
        assert_eq!(order.get_remaining_quantity(), 7);
        assert!(matches!(order.get_order_type(), OrderType::GoodTillCancel));
    }

    #[test]
    fn done_reason_strings_match_event_contract() {
        assert_eq!(DoneReason::Filled.as_str(), "filled");
        assert_eq!(DoneReason::Cancelled.as_str(), "cancelled");
        assert_eq!(DoneReason::IocExpired.as_str(), "ioc_expired");
        assert_eq!(DoneReason::FokKilled.as_str(), "fok_killed");
    }

    #[test]
    fn level_infos_expose_bids_and_asks() {
        let infos = OrderbookLevelInfos::new(
            vec![LevelInfo {
                price: 101,
                quantity: 5,
            }],
            vec![LevelInfo {
                price: 102,
                quantity: 3,
            }],
        );
        assert_eq!(infos.get_bids().len(), 1);
        assert_eq!(infos.get_asks().len(), 1);
        assert_eq!(infos.get_bids()[0].price, 101);
        assert_eq!(infos.get_asks()[0].price, 102);
    }
}
