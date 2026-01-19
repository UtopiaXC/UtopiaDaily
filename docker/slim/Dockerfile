FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY src/web/dashboard/frontend/package.json src/web/dashboard/frontend/package-lock.json* ./
RUN npm ci
COPY src/web/dashboard/frontend .
RUN npm run build

FROM python:3.9-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONPATH=/app \
    PATH="/usr/local/bin:$PATH" \
    TZ=Asia/Shanghai
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
COPY docker/entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

COPY src ./src
COPY config ./config
COPY --from=frontend-builder /app/frontend/dist ./src/web/dashboard/frontend/dist
RUN mkdir -p logs data

EXPOSE 8000

ENTRYPOINT ["entrypoint.sh"]
CMD ["python", "-m", "src.app"]