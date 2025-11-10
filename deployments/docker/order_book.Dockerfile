FROM rust:1.86-alpine AS builder

RUN apk add --no-cache musl-dev protobuf-dev

WORKDIR /build

COPY services/order_book/Cargo.toml ./services/order_book/Cargo.toml
COPY services/order_book/build.rs ./services/order_book/build.rs
COPY services/order_book/src ./services/order_book/src
COPY services/order_book/benches ./services/order_book/benches
COPY proto ./proto

WORKDIR /build/services/order_book
RUN cargo build --release --bin order_book

FROM alpine:3.18 AS runtime

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /build/services/order_book/target/release/order_book /app/order_book

EXPOSE 50056

CMD ["/app/order_book"]
