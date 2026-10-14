FROM golang:1.26.9-bookworm AS build

WORKDIR /src/webapp
COPY webapp/go.mod webapp/go.sum ./
RUN go mod download
COPY webapp/ ./

RUN mkdir -p /out/data/uploads /out/data/logs && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=readonly -trimpath -ldflags="-s -w" \
    -o /out/webshield-webapp ./cmd/server

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build --chown=65532:65532 \
    /out/webshield-webapp /webshield-webapp
COPY --from=build --chown=65532:65532 /out/data /data

USER 65532:65532
ENV WEBAPP_PORT=8081
ENV WEBAPP_DATABASE_PATH=/data/webshield.db
ENV WEBAPP_UPLOAD_DIR=/data/uploads
ENV WEBAPP_LOG_FILE=/data/logs/security_events.jsonl
ENV WEBAPP_COOKIE_SECURE=false
VOLUME ["/data"]
EXPOSE 8081
ENTRYPOINT ["/webshield-webapp"]
