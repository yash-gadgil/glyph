use crate::orderbook::types::{OrderId, TradeInfo};

#[derive(Debug, Clone)]
pub struct Trade {
    pub trade_id: String,
    pub bid_trade: TradeInfo,
    pub ask_trade: TradeInfo,
    pub taker_order_id: OrderId,
}

impl Trade {
    pub fn new(bid_trade: TradeInfo, ask_trade: TradeInfo, taker_order_id: OrderId) -> Self {
        Self {
            trade_id: uuid::Uuid::new_v4().to_string(),
            bid_trade,
            ask_trade,
            taker_order_id,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn info(id: &str) -> TradeInfo {
        TradeInfo {
            order_id: id.to_string(),
            user_id: format!("user-{}", id),
            price: 100,
            quantity: 5,
        }
    }

    #[test]
    fn trades_get_unique_ids() {
        let a = Trade::new(info("b1"), info("s1"), "b1".to_string());
        let b = Trade::new(info("b1"), info("s1"), "b1".to_string());
        assert_ne!(a.trade_id, b.trade_id);
        assert_eq!(a.taker_order_id, "b1");
        assert_eq!(a.bid_trade.user_id, "user-b1");
    }
}
