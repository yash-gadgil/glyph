use std::cmp::Reverse;
use std::collections::{BTreeMap, HashMap};
use std::rc::Rc;

use crate::{
    LevelInfo, MatchOutcome, OrderEntry, OrderId, OrderPointer, OrderPointers, OrderbookLevelInfos,
    Price, Side,
};

#[derive(Debug, Default)]
pub struct Orderbook {
    pub bids: BTreeMap<Reverse<Price>, OrderPointers>,
    pub asks: BTreeMap<Price, OrderPointers>,
    pub orders: HashMap<OrderId, OrderEntry>,
    pub last_trade_price: Option<Price>,
}

impl Orderbook {
    pub fn new() -> Self {
        Self::default()
    }

    fn knows_order(&self, id: &OrderId) -> bool {
        self.orders.contains_key(id)
    }

    pub fn add_order(&mut self, order: OrderPointer) -> MatchOutcome {
        let (id, side, price) = {
            let o = order.borrow();
            (o.get_order_id().clone(), *o.get_side(), o.get_price())
        };

        if self.knows_order(&id) {
            return MatchOutcome::rejected();
        }

        match side {
            Side::Buy => self
                .bids
                .entry(Reverse(price))
                .or_default()
                .push(Rc::clone(&order)),
            Side::Sell => self.asks.entry(price).or_default().push(Rc::clone(&order)),
        }
        self.orders.insert(id, OrderEntry { order });

        MatchOutcome {
            accepted: true,
            ..Default::default()
        }
    }

    pub fn inject_price(&mut self, price: Price) -> MatchOutcome {
        self.last_trade_price = Some(price);
        MatchOutcome::default()
    }

    pub fn get_order_infos(&self) -> OrderbookLevelInfos {
        let mut bid_infos = Vec::with_capacity(self.orders.len());
        let mut ask_infos = Vec::with_capacity(self.orders.len());

        fn create_level_info(price: Price, orders: &OrderPointers) -> LevelInfo {
            let quantity = orders
                .iter()
                .map(|o| o.borrow().get_remaining_quantity())
                .sum();
            LevelInfo { price, quantity }
        }

        for (&Reverse(price), orders) in &self.bids {
            bid_infos.push(create_level_info(price, orders));
        }

        for (&price, orders) in &self.asks {
            ask_infos.push(create_level_info(price, orders));
        }

        OrderbookLevelInfos::new(bid_infos, ask_infos)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::OrderType;
    use crate::orderbook::order::Order;
    use std::cell::RefCell;

    fn typed(order_type: OrderType, id: &str, side: Side, price: Price, qty: i64) -> OrderPointer {
        Rc::new(RefCell::new(Order::new(
            order_type,
            id.to_string(),
            format!("user-{}", id),
            side,
            price,
            qty,
        )))
    }

    fn gtc(id: &str, side: Side, price: Price, qty: i64) -> OrderPointer {
        typed(OrderType::GoodTillCancel, id, side, price, qty)
    }

    type Levels = Vec<(Price, i64)>;

    fn levels(book: &Orderbook) -> (Levels, Levels) {
        let infos = book.get_order_infos();
        let bids = infos
            .get_bids()
            .iter()
            .map(|l| (l.price, l.quantity))
            .collect();
        let asks = infos
            .get_asks()
            .iter()
            .map(|l| (l.price, l.quantity))
            .collect();
        (bids, asks)
    }

    #[test]
    fn resting_limit_produces_no_trades() {
        let mut book = Orderbook::new();
        book.inject_price(99);
        let outcome = book.add_order(gtc("b1", Side::Buy, 90, 10));
        assert!(outcome.accepted);
        assert!(outcome.trades.is_empty());
        assert!(outcome.done.is_empty());
        assert_eq!(levels(&book), (vec![(90, 10)], vec![]));
    }

    #[test]
    fn duplicate_order_id_is_rejected() {
        let mut book = Orderbook::new();
        book.add_order(gtc("b1", Side::Buy, 100, 10));
        let outcome = book.add_order(gtc("b1", Side::Buy, 105, 5));
        assert!(!outcome.accepted);
        assert_eq!(levels(&book), (vec![(100, 10)], vec![]));
    }

    #[test]
    fn level_infos_aggregate_and_sort_best_first() {
        let mut book = Orderbook::new();
        book.inject_price(105);
        book.add_order(gtc("b1", Side::Buy, 101, 3));
        book.add_order(gtc("b2", Side::Buy, 101, 7));
        book.add_order(gtc("b3", Side::Buy, 99, 5));
        book.add_order(gtc("s1", Side::Sell, 110, 6));
        book.add_order(gtc("s2", Side::Sell, 108, 4));

        let (bids, asks) = levels(&book);
        assert_eq!(bids, vec![(101, 10), (99, 5)]);
        assert_eq!(asks, vec![(108, 4), (110, 6)]);
    }
}
