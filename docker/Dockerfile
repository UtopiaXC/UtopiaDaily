# ==============================================================================
# Build Instruction:
# Run this command from the project root directory:
# docker build -f docker/Dockerfile -t utopia-daily .
# ==============================================================================

# Stage 1: Builder (Compiling dependencies)
FROM python:3.9-alpine AS builder

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
# 1. Remove bytecode (*.pyc) and cache
# 2. Remove tests directories inside packages (common in some libs)
# 3. Remove dist-info (metadata) - WARNING: This might break 'pip show' or pkg_resources, but usually fine for runtime
RUN find /install -name '*.pyc' -delete && \
    find /install -name '__pycache__' -delete && \
    find /install -name '*.exe' -delete && \
    find /install -type d -name "tests" -exec rm -rf {} + && \
    find /install -type d -name "test" -exec rm -rf {} +

# Stage 2: Final Image (Runtime only)
FROM python:3.9-alpine

WORKDIR /app

# Set PYTHONPATH to /app so imports like 'from src...' work correctly
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app
ENV PATH=/usr/local/bin:$PATH

# Install runtime dependencies
# Kept: tzdata for timezone support
RUN apk add --no-cache \
    tzdata

# Copy installed python packages from builder
COPY --from=builder /install /usr/local

# Copy application code
COPY src ./src

# Create logs directory
RUN mkdir -p logs

EXPOSE 8000

# Run module instead of script to ensure correct path handling
CMD ["python", "-m", "src.app"]
