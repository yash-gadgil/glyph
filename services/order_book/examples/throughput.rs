use std::cell::RefCell;
use std::rc::Rc;
use std::time::Instant;

use order_book::orderbook::book::Orderbook;
use order_book::orderbook::order::Order;
use order_book::{OrderType, Side};

fn make_order(id: u64, side: Side, price: i64, qty: i64) -> Rc<RefCell<Order>> {
    Rc::new(RefCell::new(Order::new(
        OrderType::GoodTillCancel,
        id.to_string(),
        format!("user-{}", id % 16),
        side,
        price,
        qty,
    )))
}

fn run(n: usize) {
    let mid: i64 = 1000;
    let workload: Vec<_> = (0..n as u64)
        .map(|i| {
            let side = if i % 2 == 0 { Side::Buy } else { Side::Sell };
            let jitter = ((i.wrapping_mul(2654435761)) % 20) as i64 - 10;
            (i + 1, side, mid + jitter)
        })
        .collect();

    let mut book = Orderbook::new();

    let start = Instant::now();
    for (id, side, price) in workload {
        let _ = book.add_order(make_order(id, side, price, 10));
    }
    let elapsed = start.elapsed();

    let secs = elapsed.as_secs_f64();
    let per_op_us = (elapsed.as_nanos() as f64) / (n as f64) / 1_000.0;
    let ops_per_sec = (n as f64) / secs;

    println!(
        "n={:>8}  total={:>10.3} ms  per_order={:>8.3} µs  throughput={:>12.0} orders/sec",
        n,
        secs * 1000.0,
        per_op_us,
        ops_per_sec
    );
}

fn main() {
    println!("warmup:");
    run(10_000);
    println!("\nresults:");
    for &n in &[10_000usize, 100_000, 500_000, 1_000_000] {
        run(n);
    }
}
