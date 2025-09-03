# Multi-stage build for combined frontend/backend architecture
FROM golang:1.24-alpine AS builder

# Install dependencies for build
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go module files first for better caching
COPY go.mod go.sum ./

# Download dependencies from remote repository
RUN go mod download

# Copy source code (all .go files for modular structure)
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o signature-tool .

# Final stage - use alpine
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Create working directory
WORKDIR /app

# Copy the built binary
COPY --from=builder /app/signature-tool .

# Copy frontend files
COPY frontend/ ./frontend/

# Make binary executable
RUN chmod +x signature-tool

# Expose port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080
ENV TEE_CONFIG_ADDR=localhost:50052
ENV FRONTEND_PATH=./frontend

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Default command
CMD ["./signature-tool"]