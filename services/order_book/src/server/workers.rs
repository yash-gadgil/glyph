use crossbeam_channel::{Receiver, Sender, unbounded};
use std::{cell::RefCell, rc::Rc, thread};

use crate::{
    orderbook::{book::Orderbook, order::Order},
    server::commands::{AddOrderResult, CancelOrderResult, Command},
};

pub fn spawn_worker_for_symbol(symbol: String) -> Sender<Command> {
    let (tx, rx): (Sender<Command>, Receiver<Command>) = unbounded();

    thread::Builder::new()
        .name(format!("ob-worker-{}", symbol))
        .spawn(move || {
            let mut book = Orderbook::new();
            worker_loop(&symbol, &mut book, rx);
        })
        .expect("failed to spawn orderbook worker thread");

    tx
}

fn worker_loop(_symbol: &str, book: &mut Orderbook, rx: Receiver<Command>) {
    while let Ok(cmd) = rx.recv() {
        match cmd {
            Command::AddOrder { order, resp } => {
                let outcome = book.add_order(Rc::new(RefCell::new(Order::new(
                    order.order_type,
                    order.id,
                    order.user_id,
                    order.side,
                    order.price,
                    order.quantity,
                ))));

                let _ = resp.send(AddOrderResult {
                    trades: outcome.trades,
                    accepted: outcome.accepted,
                    done: outcome.done,
                });
            }

            Command::CancelOrder { order_id, resp } => {
                let done = book.cancel_order(&order_id);
                let _ = resp.send(CancelOrderResult {
                    cancelled: done.is_some(),
                    done,
                });
            }
        }
    }
}
