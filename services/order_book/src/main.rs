use order_book::serve;
use std::env;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let host = env::var("SERVER_HOST").unwrap_or_else(|_| "[::1]".to_string());
    let port = env::var("SERVER_PORT").unwrap_or_else(|_| "50056".to_string());

    let addr_str = format!("{}:{}", host, port);
    println!("Orderbook server listening on {}", addr_str);
    serve(&addr_str).await
}
