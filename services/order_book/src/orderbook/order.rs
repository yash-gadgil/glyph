use crate::orderbook::types::*;

#[derive(Debug)]
pub struct Order {
    order_type: OrderType,
    order_id: OrderId,
    user_id: UserId,
    side: Side,
    price: Price,
    initial_quantity: Quantity,
    remaining_quantity: Quantity,
}

impl Order {
    pub fn new(
        order_type: OrderType,
        order_id: OrderId,
        user_id: UserId,
        side: Side,
        price: Price,
        quantity: Quantity,
    ) -> Self {
        Self {
            order_type,
            order_id,
            user_id,
            side,
            price,
            initial_quantity: quantity,
            remaining_quantity: quantity,
        }
    }

    pub fn get_order_id(&self) -> &OrderId {
        &self.order_id
    }
    pub fn get_user_id(&self) -> &UserId {
        &self.user_id
    }
    pub fn get_side(&self) -> &Side {
        &self.side
    }
    pub fn get_order_type(&self) -> &OrderType {
        &self.order_type
    }
    pub fn get_price(&self) -> Price {
        self.price
    }
    pub fn get_initial_quantity(&self) -> Quantity {
        self.initial_quantity
    }
    pub fn get_remaining_quantity(&self) -> Quantity {
        self.remaining_quantity
    }

    pub fn get_filled_quantity(&self) -> Quantity {
        self.initial_quantity - self.remaining_quantity
    }

    pub fn is_filled(&self) -> bool {
        self.remaining_quantity == 0
    }

    pub fn fill(&mut self, quantity: Quantity) -> Result<(), String> {
        if quantity > self.remaining_quantity {
            return Err(format!(
                "Order [{}] cannot be filled for more than its remaining quantity",
                self.order_id
            ));
        }
        self.remaining_quantity -= quantity;
        Ok(())
    }

    pub fn to_good_till_cancel(&mut self, price: Price) {
        self.order_type = OrderType::GoodTillCancel;
        self.price = price;
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn new_order(qty: Quantity) -> Order {
        Order::new(
            OrderType::GoodTillCancel,
            "o1".to_string(),
            "u1".to_string(),
            Side::Buy,
            100,
            qty,
        )
    }

    #[test]
    fn new_order_starts_unfilled() {
        let order = new_order(10);
        assert_eq!(order.get_order_id(), "o1");
        assert_eq!(order.get_user_id(), "u1");
        assert_eq!(order.get_price(), 100);
        assert_eq!(order.get_initial_quantity(), 10);
        assert_eq!(order.get_remaining_quantity(), 10);
        assert_eq!(order.get_filled_quantity(), 0);
        assert!(!order.is_filled());
        assert!(matches!(order.get_side(), Side::Buy));
        assert!(matches!(order.get_order_type(), OrderType::GoodTillCancel));
    }

    #[test]
    fn partial_fill_reduces_remaining() {
        let mut order = new_order(10);
        order.fill(4).unwrap();
        assert_eq!(order.get_remaining_quantity(), 6);
        assert_eq!(order.get_filled_quantity(), 4);
        assert!(!order.is_filled());
    }

    #[test]
    fn full_fill_marks_order_filled() {
        let mut order = new_order(10);
        order.fill(10).unwrap();
        assert!(order.is_filled());
    }

    #[test]
    fn sequential_fills_accumulate() {
        let mut order = new_order(10);
        order.fill(3).unwrap();
        order.fill(3).unwrap();
        order.fill(4).unwrap();
        assert!(order.is_filled());
    }

    #[test]
    fn overfill_is_rejected_and_leaves_order_unchanged() {
        let mut order = new_order(5);
        let err = order.fill(6).unwrap_err();
        assert!(err.contains("o1"));
        assert_eq!(order.get_remaining_quantity(), 5);
    }

    #[test]
    fn zero_quantity_fill_is_a_noop() {
        let mut order = new_order(5);
        order.fill(0).unwrap();
        assert_eq!(order.get_remaining_quantity(), 5);
    }

    #[test]
    fn market_order_converts_to_good_till_cancel() {
        let mut order = Order::new(
            OrderType::Market,
            "m1".to_string(),
            "u1".to_string(),
            Side::Sell,
            0,
            5,
        );
        order.to_good_till_cancel(250);
        assert!(matches!(order.get_order_type(), OrderType::GoodTillCancel));
        assert_eq!(order.get_price(), 250);
        assert_eq!(order.get_remaining_quantity(), 5);
    }
}
