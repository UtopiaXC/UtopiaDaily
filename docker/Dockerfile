# ==============================================================================
# Build Instruction:
# Run this command from the project root directory:
# docker build -f docker/Dockerfile -t utopia-daily .
# ==============================================================================

# Stage 1: Frontend Builder
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend source code
# 注意：这里假设您的前端代码在 src/web/dashboard/frontend
COPY src/web/dashboard/frontend/package.json src/web/dashboard/frontend/package-lock.json* ./
RUN npm install

COPY src/web/dashboard/frontend .
# 构建生产版本，输出到 /app/frontend/dist
RUN npm run build

# Stage 2: Backend Builder (Compiling dependencies)
FROM python:3.9-alpine AS backend-builder

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Install minimal build dependencies
RUN apk add --no-cache \
    gcc \
    musl-dev \
    libffi-dev

COPY requirements.txt .

# Install dependencies to a specific prefix
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# Aggressive cleanup to reduce image size
RUN find /install -name '*.pyc' -delete && \
    find /install -name '__pycache__' -delete && \
    find /install -name '*.exe' -delete && \
    find /install -type d -name "tests" -exec rm -rf {} + && \
    find /install -type d -name "test" -exec rm -rf {} +

# Stage 3: Final Image (Runtime only)
FROM python:3.9-alpine

WORKDIR /app

# Set PYTHONPATH to /app so imports like 'from src...' work correctly
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app
ENV PATH=/usr/local/bin:$PATH

# Install runtime dependencies
RUN apk add --no-cache \
    tzdata

# Copy installed python packages from backend-builder
COPY --from=backend-builder /install /usr/local

# Copy application code
COPY src ./src

# Copy compiled frontend assets from frontend-builder
# 将构建好的 dist 目录放置在 src/web/dashboard/frontend/dist
# 这样我们只需要修改 entry.py 指向这个 dist 目录即可
COPY --from=frontend-builder /app/frontend/dist ./src/web/dashboard/frontend/dist

# Create logs directory
RUN mkdir -p logs

EXPOSE 8000

# Run module instead of script to ensure correct path handling
CMD ["python", "-m", "src.app"]
