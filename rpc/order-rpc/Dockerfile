FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY rpc/order-rpc/order-rpc .
COPY rpc/order-rpc/etc/order.yaml etc/

EXPOSE 8081

ENTRYPOINT ["./order-rpc", "-f", "etc/order.yaml"]