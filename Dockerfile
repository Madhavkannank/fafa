# Production Dockerfile for JSBI Go Port
# Stage 1: Build & Verification Environment
FROM golang:1.22-alpine

# Set working directory
WORKDIR /app

# Copy dependency manifests and source code
COPY . .

# Run test suite by default
CMD ["go", "test", "-v", "./tests/port/..."]
