pub mod orderbook;
pub use orderbook::types::*;
pub mod server;

use std::env;
use std::net::SocketAddr;
use tonic::transport::Server;

pub async fn serve(addr: &str) -> Result<(), Box<dyn std::error::Error>> {
    let socket: SocketAddr = addr.parse()?;

    let rmq_channel = match env::var("RABBITMQ_URL") {
        Ok(url) => {
            println!("Connecting to RabbitMQ at {}", url);
            match lapin::Connection::connect(&url, lapin::ConnectionProperties::default()).await {
                Ok(conn) => {
                    let ch = conn.create_channel().await?;
                    ch.exchange_declare(
                        "order.events",
                        lapin::ExchangeKind::Direct,
                        lapin::options::ExchangeDeclareOptions {
                            durable: true,
                            ..Default::default()
                        },
                        lapin::types::FieldTable::default(),
                    )
                    .await?;
                    println!("Connected to RabbitMQ");
                    Some(ch)
                }
                Err(e) => {
                    eprintln!(
                        "RabbitMQ connect failed: {}. Running without publishing.",
                        e
                    );
                    None
                }
            }
        }
        Err(_) => {
            println!("RABBITMQ_URL not set. Running without publishing.");
            None
        }
    };

    let svc = server::OrderBookServer::new(rmq_channel);

    Server::builder()
        .add_service(
            server::order_book_service::orderbook_service_server::OrderbookServiceServer::new(svc),
        )
        .serve(socket)
        .await?;

    Ok(())
}
