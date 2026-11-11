FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend-builder

ARG APP_VERSION=0.2.0
ENV APP_VERSION=${APP_VERSION}

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

FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS backend-builder

ENV GOTOOLCHAIN=auto
ARG TARGETOS
ARG TARGETARCH
ARG APP_VERSION=0.2.0

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY web/embed.go ./web/embed.go
COPY --from=frontend-builder /src/web/dashboard/dist ./web/dashboard/dist
COPY --from=frontend-builder /src/web/newspaper/dist ./web/newspaper/dist

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -trimpath \
    -ldflags="-s -w -X github.com/UtopiaXC/UtopiaDaily/internal/config.Version=${APP_VERSION} -X github.com/UtopiaXC/UtopiaDaily/internal/module.DefaultVersion=${APP_VERSION}" \
    -o /out/utopia_daily ./cmd/utopia_daily

FROM debian:bookworm-slim

ARG APP_VERSION=0.2.0

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

RUN APP_VERSION="${APP_VERSION}" python3 -c "import json, os, re; mp='/app/plugins/scrapers/example_hn_rss/manifest.json'; py='/app/plugins/scrapers/example_hn_rss/main.py'; v=os.environ.get('APP_VERSION','0.2.0'); (lambda d: json.dump({**d, 'version': v}, open(mp, 'w'), indent=2))(json.load(open(mp))) if os.path.exists(mp) else None; open(py, 'w').write(re.sub(r'UtopiaDaily/[0-9A-Za-z._-]+', 'UtopiaDaily/' + v, open(py).read())) if os.path.exists(py) else None" \
    && mkdir -p \
    /app/data/config \
    /app/data/db \
    /app/cache/locales \
    /app/logs \
    /app/plugins/scrapers \
    /app/plugins/push

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD ["python3", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/healthz', timeout=3)"]

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/app/utopia_daily"]
 