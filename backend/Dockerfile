FROM golang:1.26.9-bookworm AS build

WORKDIR /src/backend
COPY backend/go.mod ./
COPY backend/ ./

RUN mkdir -p /out && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=readonly -trimpath -ldflags="-s -w" \
    -o /out/webshield-backend ./cmd/server

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build --chown=65532:65532 \
    /out/webshield-backend /webshield-backend

USER 65532:65532
ENV HTTP_ADDR=0.0.0.0:8080
EXPOSE 8080
ENTRYPOINT ["/webshield-backend"]
