FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY rpc/stock-rpc/stock-rpc .
COPY rpc/stock-rpc/etc/stock.yaml etc/

EXPOSE 8082

ENTRYPOINT ["./stock-rpc", "-f", "etc/stock.yaml"]