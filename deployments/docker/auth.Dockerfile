FROM golang:1.25.1-alpine AS builder

WORKDIR /build

COPY services/auth .
COPY services/gen/golang /gen/golang

#RUN go mod download

RUN go build -mod=vendor -o ./auth

FROM alpine:3.18 AS runtime

WORKDIR /app
COPY --from=builder /build/auth /app/auth

CMD ["/app/auth"]