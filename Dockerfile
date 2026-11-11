FROM node:22-alpine AS frontend-builder

WORKDIR /src/web/dashboard
COPY web/dashboard/package.json web/dashboard/package-lock.json* ./
RUN npm ci
COPY web/dashboard .
RUN npm run build

WORKDIR /src/web/newspaper
COPY web/newspaper/package.json web/newspaper/package-lock.json* ./
RUN npm ci
COPY web/newspaper .
RUN npm run build

FROM golang:1.25-bookworm AS backend-builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY web/embed.go ./web/embed.go
COPY --from=frontend-builder /src/web/dashboard/dist ./web/dashboard/dist
COPY --from=frontend-builder /src/web/newspaper/dist ./web/newspaper/dist

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/utopia_daily ./cmd/utopia_daily

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    TZ=Asia/Shanghai \
    UDAILY_SERVER_HOST=0.0.0.0 \
    UDAILY_SERVER_PORT=8000 \
    UDAILY_SERVER_MODE=release

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    nodejs \
    python3 \
    python3-pip \
    python3-venv \
    tini \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /usr/local/bin/

WORKDIR /app

COPY --from=backend-builder /out/utopia_daily /app/utopia_daily
COPY data/config/config.example.yaml /app/data/config/config.example.yaml
COPY plugins /app/plugins

RUN mkdir -p \
    /app/data/config \
    /app/data/db \
    /app/cache/locales \
    /app/logs \
    /app/plugins/scrapers \
    /app/plugins/push

EXPOSE 8000

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/app/utopia_daily"]
