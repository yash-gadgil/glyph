FROM golang:1.25.1-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ pkg/
COPY services/gen/ services/gen/
COPY services/advisor/ services/advisor/

RUN CGO_ENABLED=0 go build -o /out/advisor ./services/advisor

FROM alpine:3.18 AS runtime

WORKDIR /app
COPY --from=builder /out/advisor /app/advisor

CMD ["/app/advisor"]
