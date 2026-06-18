# Build Stage
FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary statically, without cgo
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /osto-auth-cli ./cmd/osto-auth-cli

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /osto-auth-cli /app/osto-auth-cli

# Create a data directory for the SQLite database
RUN mkdir -p /data

# Default command
ENTRYPOINT ["/app/osto-auth-cli"]
