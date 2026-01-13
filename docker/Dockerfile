# ==============================================================================
# Build Instruction:
# Run this command from the project root directory (not inside /docker):
# docker build -f docker/Dockerfile -t utopia-daily .
# ==============================================================================

# Stage 1: Builder
FROM python:3.10-slim as builder

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --user --no-cache-dir -r requirements.txt

# Stage 2: Final Image
FROM python:3.10-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app
ENV PATH=/root/.local/bin:$PATH

# Copy only necessary files from builder
COPY --from=builder /root/.local /root/.local

# Copy application code
COPY src ./src

# Create logs directory
RUN mkdir -p logs

EXPOSE 8000

CMD ["python", "src/app.py"]
