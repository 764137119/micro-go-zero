FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY api-gateway/api-gateway ./
COPY api-gateway/etc/config.yaml ./etc/

EXPOSE 8888

ENTRYPOINT ["./api-gateway"]