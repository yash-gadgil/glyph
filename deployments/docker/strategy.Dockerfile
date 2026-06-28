FROM golang:1.25.1-alpine AS builder

ENV GOTOOLCHAIN=local

WORKDIR /build

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.0

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ pkg/
COPY services/gen/ services/gen/
COPY services/strategy/ services/strategy/

RUN CGO_ENABLED=0 go build -o /out/strategy ./services/strategy

FROM alpine:3.18 AS runtime

WORKDIR /app
COPY --from=builder /out/strategy /app/strategy
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /build/services/strategy/db/migrations /app/db/migrations

CMD ["/app/strategy"]
