FROM python:3.10-slim

# Version and metadata
LABEL version="0.0.6"
LABEL description="Raqeem Devices Backend - FastAPI IoT device monitoring service"

# Set working directory
WORKDIR /app

# Install build dependencies (optional but helps for packages with native extensions)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy app code
COPY src/ .

# Start FastAPI app with Uvicorn (PORT must be provided via env)
CMD ["sh", "-lc", "test -n \"$PORT\" || { echo 'ERROR: PORT env is required'; exit 1; }; uvicorn app.main:app --host 0.0.0.0 --port \"$PORT\""]
