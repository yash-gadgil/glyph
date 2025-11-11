use std::cell::RefCell;
use std::rc::Rc;
use std::time::Instant;

use criterion::{Criterion, Throughput, criterion_group, criterion_main};
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

fn build_workload(n: usize) -> Vec<Rc<RefCell<Order>>> {
    let mid: i64 = 1000;
    (0..n as u64)
        .map(|i| {
            let side = if i % 2 == 0 { Side::Buy } else { Side::Sell };
            let jitter = ((i.wrapping_mul(2654435761)) % 20) as i64 - 10;
            let price = mid + jitter;
            make_order(i + 1, side, price, 10)
        })
        .collect()
}

fn bench_add_order(c: &mut Criterion) {
    let mut group = c.benchmark_group("orderbook");

    for &n in &[1_000usize, 10_000usize] {
        group.throughput(Throughput::Elements(n as u64));
        group.bench_function(format!("add_order_mixed/{}", n), |b| {
            b.iter_custom(|iters| {
                let mut total = std::time::Duration::ZERO;
                for _ in 0..iters {
                    let workload = build_workload(n);
                    let mut book = Orderbook::new();
                    let start = Instant::now();
                    for order in workload {
                        let _ = book.add_order(order);
                    }
                    total += start.elapsed();
                }
                total
            })
        });
    }

    group.finish();
}

criterion_group!(benches, bench_add_order);
criterion_main!(benches);
