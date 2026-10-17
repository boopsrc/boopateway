# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
# modernc.org/sqlite é Go puro: binário estático, sem CGO.
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/portal ./cmd/portal
RUN mkdir -p /out/data/images

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
COPY --from=build /out/portal /portal
COPY --from=build --chown=nonroot:nonroot /out/data /data
ENV ADDR=:8080 \
    DB_PATH=/data/boopateway.db \
    IMAGE_DIR=/data/images \
    GIN_MODE=release
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["/api"]
