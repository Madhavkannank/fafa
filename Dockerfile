# -----------------------------------------------------------------------------
# Stage 1 — Build & Verification
# -----------------------------------------------------------------------------
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install make and git for repository build tooling
RUN apk add --no-cache make git

# Copy module files first for layer caching
COPY go.mod go.sum* ./

# Download dependencies (if any)
RUN go mod download || true

# Copy repository content
COPY . .

# Verify library compilation
RUN go build -v ./src/...

# Verify test suite execution
RUN go test -v ./tests/port/...

# -----------------------------------------------------------------------------
# Stage 2 — Submission Runtime Container
# -----------------------------------------------------------------------------
FROM golang:1.22-alpine

WORKDIR /app

# Install make and git for runtime verification commands
RUN apk add --no-cache make git

# Copy prepared repository from builder stage
COPY --from=builder /app /app

# Default command executes full test suite verification
CMD ["go", "test", "-v", "./tests/port/..."]
