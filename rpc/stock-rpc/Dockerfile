FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o stock-rpc .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /build/stock-rpc .
COPY --from=builder /build/etc/stock.yaml etc/

EXPOSE 8082

ENTRYPOINT ["./stock-rpc", "-f", "etc/stock.yaml"]