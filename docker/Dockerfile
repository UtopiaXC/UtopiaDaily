# docker build -f docker/Dockerfile -t utopia-daily .

FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

COPY src/web/dashboard/frontend/package.json src/web/dashboard/frontend/package-lock.json* ./
RUN npm install

COPY src/web/dashboard/frontend .
RUN npm run build

FROM python:3.9-slim AS backend-builder

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

RUN apt-get update && apt-get install -y \
    gcc \
    libpq-dev \
    libjpeg-dev \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

RUN find /install -name '*.pyc' -delete && \
    find /install -name '__pycache__' -delete && \
    find /install -name '*.exe' -delete && \
    find /install -type d -name "tests" -exec rm -rf {} + && \
    find /install -type d -name "test" -exec rm -rf {} +

FROM python:3.9-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app
ENV PATH=/usr/local/bin:$PATH

RUN apt-get update && apt-get install -y \
    libpq5 \
    libjpeg62-turbo \
    zlib1g \
    tzdata \
    fonts-dejavu \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend-builder /install /usr/local

COPY src ./src
COPY config ./config

COPY --from=frontend-builder /app/frontend/dist ./src/web/dashboard/frontend/dist

RUN mkdir -p logs

EXPOSE 8000

CMD ["python", "-m", "src.app"]
