FROM golang:1.25.1-alpine AS builder

WORKDIR /build

COPY gateway .
COPY services/gen/golang /services/gen/golang

RUN go mod download

RUN go build -o ./gateway

FROM alpine:3.18 AS runtime

WORKDIR /app
COPY --from=builder /build/gateway /app/gateway
COPY .env /app/.env

CMD ["/app/gateway"]