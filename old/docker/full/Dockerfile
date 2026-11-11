FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY src/web/dashboard/frontend/package.json src/web/dashboard/frontend/package-lock.json* ./
RUN npm ci
COPY src/web/dashboard/frontend .
RUN npm run build


FROM python:3.9-slim AS model-downloader
WORKDIR /tmp
RUN pip install huggingface_hub
COPY docker/full/preload_models.py .
RUN python preload_models.py


FROM python:3.9-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONPATH=/app \
    PATH="/usr/local/bin:$PATH" \
    TZ=Asia/Shanghai

RUN apt-get update && apt-get install -y --no-install-recommends \
    libjpeg62-turbo \
    zlib1g \
    libpq5 \
    tzdata \
    ca-certificates \
    curl \
    libgomp1 \
    fonts-dejavu \
    && ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir --compile -r requirements.txt
COPY src ./src
COPY config ./config
COPY --from=frontend-builder /app/frontend/dist ./src/web/dashboard/frontend/dist
COPY --from=model-downloader /app/models ./models
RUN mkdir -p logs data

EXPOSE 8000

CMD ["python", "-m", "src.app"]