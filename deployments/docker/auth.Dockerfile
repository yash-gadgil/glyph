FROM golang:1.25.1-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ pkg/
COPY services/gen/ services/gen/
COPY services/auth/ services/auth/

RUN CGO_ENABLED=0 go build -o /out/auth ./services/auth

FROM alpine:3.18 AS runtime

WORKDIR /app
COPY --from=builder /out/auth /app/auth

CMD ["/app/auth"]
