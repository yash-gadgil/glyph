use std::cmp::Reverse;
use std::collections::{BTreeMap, HashMap};
use std::rc::Rc;

use crate::orderbook::trade::Trade;
use crate::{
    DoneInfo, DoneReason, LevelInfo, MatchOutcome, OrderEntry, OrderId, OrderModify, OrderPointer,
    OrderPointers, OrderType, OrderbookLevelInfos, Price, Side, TradeInfo,
};

pub const MARKET_COUNTERPARTY: &str = "market";

#[derive(Debug, Default)]
pub struct Orderbook {
    pub bids: BTreeMap<Reverse<Price>, OrderPointers>,
    pub asks: BTreeMap<Price, OrderPointers>,
    pub orders: HashMap<OrderId, OrderEntry>,
    pub pending_markets: Vec<OrderPointer>,
    pub last_trade_price: Option<Price>,
}

fn marketable(side: &Side, limit: Price, market: Price) -> bool {
    match side {
        Side::Buy => limit >= market,
        Side::Sell => limit <= market,
    }
}

impl Orderbook {
    pub fn new() -> Self {
        Self::default()
    }

    fn knows_order(&self, id: &OrderId) -> bool {
        self.orders.contains_key(id)
            || self
                .pending_markets
                .iter()
                .any(|o| o.borrow().get_order_id() == id)
    }

    fn fill_against_market(
        &mut self,
        order: &OrderPointer,
        price: Price,
        outcome: &mut MatchOutcome,
    ) {
        let (id, user, side, qty) = {
            let o = order.borrow();
            (
                o.get_order_id().clone(),
                o.get_user_id().clone(),
                *o.get_side(),
                o.get_remaining_quantity(),
            )
        };
        if qty == 0 {
            return;
        }
        if let Err(msg) = order.borrow_mut().fill(qty) {
            eprintln!("{}", msg);
            return;
        }

        let user_info = TradeInfo {
            order_id: id.clone(),
            user_id: user.clone(),
            price,
            quantity: qty,
        };
        let market_info = TradeInfo {
            order_id: MARKET_COUNTERPARTY.to_string(),
            user_id: MARKET_COUNTERPARTY.to_string(),
            price,
            quantity: qty,
        };

        let trade = match side {
            Side::Buy => Trade::new(user_info, market_info, id.clone()),
            Side::Sell => Trade::new(market_info, user_info, id.clone()),
        };
        outcome.trades.push(trade);

        self.orders.remove(&id);
        outcome.done.push(DoneInfo {
            order_id: id,
            user_id: user,
            reason: DoneReason::Filled,
            unfilled_qty: 0,
        });
    }

    pub fn add_order(&mut self, order: OrderPointer) -> MatchOutcome {
        if self.knows_order(order.borrow().get_order_id()) {
            return MatchOutcome::rejected();
        }

        let mut outcome = MatchOutcome {
            accepted: true,
            ..Default::default()
        };

        let (order_id, user_id, price, side, qty, order_type) = {
            let o = order.borrow();
            (
                o.get_order_id().clone(),
                o.get_user_id().clone(),
                o.get_price(),
                *o.get_side(),
                o.get_remaining_quantity(),
                *o.get_order_type(),
            )
        };

        match order_type {
            OrderType::Market => match self.last_trade_price {
                Some(market) => self.fill_against_market(&order, market, &mut outcome),
                None => self.pending_markets.push(order),
            },

            OrderType::FillAndKill | OrderType::ImmediateOrCancel | OrderType::FillOrKill => {
                match self.last_trade_price {
                    Some(market) if marketable(&side, price, market) => {
                        self.fill_against_market(&order, market, &mut outcome);
                    }
                    _ => {
                        let reason = if matches!(order_type, OrderType::FillOrKill) {
                            DoneReason::FokKilled
                        } else {
                            DoneReason::IocExpired
                        };
                        outcome.done.push(DoneInfo {
                            order_id,
                            user_id,
                            reason,
                            unfilled_qty: qty,
                        });
                    }
                }
            }

            OrderType::GoodTillCancel => {
                if let Some(market) = self.last_trade_price {
                    if marketable(&side, price, market) {
                        self.fill_against_market(&order, market, &mut outcome);
                        return outcome;
                    }
                }
                if matches!(side, Side::Buy) {
                    self.bids
                        .entry(Reverse(price))
                        .or_default()
                        .push(order.clone());
                } else {
                    self.asks.entry(price).or_default().push(order.clone());
                }
                self.orders.insert(order_id, OrderEntry { order });
            }
        }

        outcome
    }

    pub fn inject_price(&mut self, price: Price) -> MatchOutcome {
        self.last_trade_price = Some(price);

        let mut outcome = MatchOutcome {
            accepted: true,
            ..Default::default()
        };

        let pending = std::mem::take(&mut self.pending_markets);
        for order in pending {
            self.fill_against_market(&order, price, &mut outcome);
        }

        loop {
            let Some((&Reverse(bid_price), _)) = self.bids.first_key_value() else {
                break;
            };
            if bid_price < price {
                break;
            }
            let orders = self.bids.remove(&Reverse(bid_price)).unwrap_or_default();
            for order in orders {
                self.fill_against_market(&order, price, &mut outcome);
            }
        }

        loop {
            let Some((&ask_price, _)) = self.asks.first_key_value() else {
                break;
            };
            if ask_price > price {
                break;
            }
            let orders = self.asks.remove(&ask_price).unwrap_or_default();
            for order in orders {
                self.fill_against_market(&order, price, &mut outcome);
            }
        }

        outcome
    }

    pub fn cancel_order(&mut self, order_id: &OrderId) -> Option<DoneInfo> {
        if let Some(idx) = self
            .pending_markets
            .iter()
            .position(|o| o.borrow().get_order_id() == order_id)
        {
            let order = self.pending_markets.remove(idx);
            let o = order.borrow();
            return Some(DoneInfo {
                order_id: o.get_order_id().clone(),
                user_id: o.get_user_id().clone(),
                reason: DoneReason::Cancelled,
                unfilled_qty: o.get_remaining_quantity(),
            });
        }

        let entry = self.orders.remove(order_id)?;
        let order_ptr = entry.order;

        let side = *order_ptr.borrow().get_side();
        let price = order_ptr.borrow().get_price();
        let remaining = order_ptr.borrow().get_remaining_quantity();
        let user_id = order_ptr.borrow().get_user_id().clone();

        match side {
            Side::Sell => {
                if let Some(price_entry) = self.asks.get_mut(&price) {
                    if let Some(idx) = price_entry.iter().position(|o| Rc::ptr_eq(o, &order_ptr)) {
                        price_entry.remove(idx);
                    }
                    if price_entry.is_empty() {
                        self.asks.remove(&price);
                    }
                }
            }
            Side::Buy => {
                let rev = Reverse(price);
                if let Some(price_entry) = self.bids.get_mut(&rev) {
                    if let Some(idx) = price_entry.iter().position(|o| Rc::ptr_eq(o, &order_ptr)) {
                        price_entry.remove(idx);
                    }
                    if price_entry.is_empty() {
                        self.bids.remove(&rev);
                    }
                }
            }
        }

        Some(DoneInfo {
            order_id: order_id.clone(),
            user_id,
            reason: DoneReason::Cancelled,
            unfilled_qty: remaining,
        })
    }

    pub fn modify_order(&mut self, order: OrderModify) -> MatchOutcome {
        let order_type = {
            let existing_entry = match self.orders.get(order.get_order_id()) {
                Some(entry) => entry,
                None => return MatchOutcome::rejected(),
            };
            *existing_entry.order.borrow().get_order_type()
        };

        self.cancel_order(order.get_order_id());
        self.add_order(order.to_order_pointer(order_type))
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
    use std::rc::Rc;

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

    fn market(id: &str, side: Side, qty: i64) -> OrderPointer {
        typed(OrderType::Market, id, side, 0, qty)
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

    fn done_reasons(outcome: &MatchOutcome) -> Vec<(&str, &str)> {
        outcome
            .done
            .iter()
            .map(|d| (d.order_id.as_str(), d.reason.as_str()))
            .collect()
    }

    fn assert_all_market_fills(outcome: &MatchOutcome) {
        for t in &outcome.trades {
            let market_sides = [&t.bid_trade, &t.ask_trade]
                .iter()
                .filter(|i| i.order_id == MARKET_COUNTERPARTY)
                .count();
            assert_eq!(market_sides, 1, "trade must have exactly one market side");
        }
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
    fn duplicate_pending_market_id_is_rejected() {
        let mut book = Orderbook::new();
        book.add_order(market("m1", Side::Buy, 5));
        assert!(!book.add_order(market("m1", Side::Buy, 5)).accepted);
    }

    #[test]
    fn market_order_fills_immediately_at_last_price() {
        let mut book = Orderbook::new();
        book.inject_price(12_345);

        let outcome = book.add_order(market("m1", Side::Buy, 8));
        assert_eq!(outcome.trades.len(), 1);
        assert_eq!(outcome.trades[0].bid_trade.order_id, "m1");
        assert_eq!(outcome.trades[0].bid_trade.price, 12_345);
        assert_eq!(outcome.trades[0].bid_trade.quantity, 8);
        assert_eq!(outcome.trades[0].ask_trade.order_id, MARKET_COUNTERPARTY);
        assert_eq!(done_reasons(&outcome), vec![("m1", "filled")]);
        assert_all_market_fills(&outcome);
        assert_eq!(levels(&book), (vec![], vec![]));
    }

    #[test]
    fn market_order_with_no_price_waits_for_first_tick() {
        let mut book = Orderbook::new();
        let outcome = book.add_order(market("m1", Side::Sell, 5));
        assert!(outcome.accepted);
        assert!(outcome.trades.is_empty());

        let outcome = book.inject_price(10_000);
        assert_eq!(outcome.trades.len(), 1);
        assert_eq!(outcome.trades[0].ask_trade.order_id, "m1");
        assert_eq!(outcome.trades[0].ask_trade.price, 10_000);
        assert!(book.pending_markets.is_empty());
    }

    #[test]
    fn marketable_buy_limit_fills_at_market_price_not_limit() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);

        let outcome = book.add_order(gtc("b1", Side::Buy, 10_500, 10));
        assert_eq!(outcome.trades.len(), 1);
        assert_eq!(outcome.trades[0].bid_trade.price, 10_000);
        assert_eq!(done_reasons(&outcome), vec![("b1", "filled")]);
    }

    #[test]
    fn resting_limit_fills_when_price_crosses() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);

        book.add_order(gtc("b1", Side::Buy, 9_500, 10));
        assert_eq!(levels(&book).0, vec![(9_500, 10)]);

        let outcome = book.inject_price(9_400);
        assert_eq!(outcome.trades.len(), 1);
        assert_eq!(outcome.trades[0].bid_trade.order_id, "b1");
        assert_eq!(outcome.trades[0].bid_trade.price, 9_400);
        assert_eq!(levels(&book), (vec![], vec![]));
    }

    #[test]
    fn sell_limit_fills_when_price_rises_through() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);

        book.add_order(gtc("s1", Side::Sell, 10_500, 4));
        let outcome = book.inject_price(10_600);
        assert_eq!(outcome.trades.len(), 1);
        assert_eq!(outcome.trades[0].ask_trade.order_id, "s1");
        assert_eq!(outcome.trades[0].ask_trade.price, 10_600);
    }

    #[test]
    fn uncrossed_limits_stay_resting_across_ticks() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);
        book.add_order(gtc("b1", Side::Buy, 9_000, 10));
        book.add_order(gtc("s1", Side::Sell, 11_000, 10));

        let outcome = book.inject_price(10_050);
        assert!(outcome.trades.is_empty());
        assert_eq!(levels(&book), (vec![(9_000, 10)], vec![(11_000, 10)]));
    }

    #[test]
    fn users_never_trade_with_each_other() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);

        book.add_order(gtc("b1", Side::Buy, 9_800, 10));
        book.add_order(gtc("s1", Side::Sell, 10_200, 10));

        let outcome = book.inject_price(9_700);
        assert_eq!(outcome.trades.len(), 1);
        assert_eq!(outcome.trades[0].bid_trade.order_id, "b1");
        assert_all_market_fills(&outcome);

        let outcome = book.inject_price(10_300);
        assert_eq!(outcome.trades.len(), 1);
        assert_eq!(outcome.trades[0].ask_trade.order_id, "s1");
        assert_all_market_fills(&outcome);
    }

    #[test]
    fn ioc_killed_when_not_marketable() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);
        let outcome = book.add_order(typed(
            OrderType::ImmediateOrCancel,
            "i1",
            Side::Buy,
            9_000,
            5,
        ));

        assert!(outcome.accepted);
        assert!(outcome.trades.is_empty());
        assert_eq!(done_reasons(&outcome), vec![("i1", "ioc_expired")]);
        assert_eq!(outcome.done[0].unfilled_qty, 5);
        assert_eq!(levels(&book), (vec![], vec![]));
    }

    #[test]
    fn ioc_fills_fully_when_marketable() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);
        let outcome = book.add_order(typed(
            OrderType::ImmediateOrCancel,
            "i1",
            Side::Buy,
            10_000,
            5,
        ));

        assert_eq!(outcome.trades.len(), 1);
        assert_eq!(outcome.trades[0].bid_trade.quantity, 5);
        assert_eq!(done_reasons(&outcome), vec![("i1", "filled")]);
    }

    #[test]
    fn fok_kills_when_not_marketable() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);
        let outcome = book.add_order(typed(OrderType::FillOrKill, "f1", Side::Sell, 11_000, 10));

        assert!(outcome.trades.is_empty());
        assert_eq!(done_reasons(&outcome), vec![("f1", "fok_killed")]);
        assert_eq!(outcome.done[0].unfilled_qty, 10);
    }

    #[test]
    fn fok_fills_fully_when_marketable() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);
        let outcome = book.add_order(typed(OrderType::FillOrKill, "f1", Side::Buy, 10_100, 10));

        let total: i64 = outcome.trades.iter().map(|t| t.bid_trade.quantity).sum();
        assert_eq!(total, 10);
        assert_eq!(done_reasons(&outcome), vec![("f1", "filled")]);
    }

    #[test]
    fn cancel_resting_order_reports_remaining_qty() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);
        book.add_order(gtc("b1", Side::Buy, 9_000, 10));

        let done = book.cancel_order(&"b1".to_string()).unwrap();
        assert_eq!(done.reason, DoneReason::Cancelled);
        assert_eq!(done.unfilled_qty, 10);
        assert_eq!(done.user_id, "user-b1");
        assert_eq!(levels(&book), (vec![], vec![]));

        assert!(book.cancel_order(&"b1".to_string()).is_none());
    }

    #[test]
    fn cancel_pending_market_order() {
        let mut book = Orderbook::new();
        book.add_order(market("m1", Side::Buy, 7));

        let done = book.cancel_order(&"m1".to_string()).unwrap();
        assert_eq!(done.reason, DoneReason::Cancelled);
        assert_eq!(done.unfilled_qty, 7);

        let outcome = book.inject_price(10_000);
        assert!(outcome.trades.is_empty());
    }

    #[test]
    fn cancel_unknown_order_returns_none() {
        let mut book = Orderbook::new();
        assert!(book.cancel_order(&"ghost".to_string()).is_none());
    }

    #[test]
    fn modify_order_moves_price_and_can_fill() {
        let mut book = Orderbook::new();
        book.inject_price(10_000);
        book.add_order(gtc("b1", Side::Buy, 9_000, 5));

        let outcome = book.modify_order(OrderModify::new(
            "b1".to_string(),
            "user-b1".to_string(),
            Side::Buy,
            10_100,
            5,
        ));
        assert_eq!(outcome.trades.len(), 1);
        assert_eq!(outcome.trades[0].bid_trade.price, 10_000);
        assert_eq!(levels(&book), (vec![], vec![]));
    }

    #[test]
    fn modify_unknown_order_is_rejected() {
        let mut book = Orderbook::new();
        let outcome = book.modify_order(OrderModify::new(
            "nope".to_string(),
            "u".to_string(),
            Side::Buy,
            105,
            5,
        ));
        assert!(!outcome.accepted);
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

    #[test]
    fn fills_conserve_quantities() {
        let mut book = Orderbook::new();
        book.inject_price(1_000);

        let mut placed = 0i64;
        let mut outcomes = Vec::new();
        for i in 0..100u64 {
            let side = if i % 2 == 0 { Side::Buy } else { Side::Sell };
            let jitter = ((i.wrapping_mul(2654435761)) % 21) as i64 - 10;
            let qty = 1 + (i % 7) as i64;
            placed += qty;
            outcomes.push(book.add_order(typed(
                OrderType::GoodTillCancel,
                &format!("o{}", i),
                side,
                1_000 + jitter,
                qty,
            )));
        }
        outcomes.push(book.inject_price(900));
        outcomes.push(book.inject_price(1_100));

        let mut filled = 0i64;
        for outcome in &outcomes {
            for t in &outcome.trades {
                assert_eq!(t.bid_trade.quantity, t.ask_trade.quantity);
                assert!(t.bid_trade.quantity > 0);
                filled += t.bid_trade.quantity;
            }
            assert_all_market_fills(outcome);
        }

        assert_eq!(placed, filled, "every placed share fills exactly once");
        assert_eq!(levels(&book), (vec![], vec![]));
    }
}
