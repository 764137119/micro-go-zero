FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o order-rpc .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /build/order-rpc .
COPY --from=builder /build/etc/order.yaml etc/

EXPOSE 8081

ENTRYPOINT ["./order-rpc", "-f", "etc/order.yaml"]