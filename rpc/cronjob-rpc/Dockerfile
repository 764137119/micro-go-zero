FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY rpc/cronjob-rpc/cronjob-rpc .
COPY rpc/cronjob-rpc/etc/cronjob.yaml etc/

EXPOSE 8083

ENTRYPOINT ["./cronjob-rpc", "-f", "etc/cronjob.yaml"]