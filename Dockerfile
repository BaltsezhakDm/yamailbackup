# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Switch to http to avoid TLS issues with Alpine mirrors
RUN sed -i 's/https/http/g' /etc/apk/repositories

# Install build dependencies if needed
RUN apk add --no-cache git

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/main.go

# Stage 2: Create a minimal runtime image
FROM alpine:3.21

WORKDIR /app

# Switch to http to avoid TLS issues with Alpine mirrors
RUN sed -i 's/https/http/g' /etc/apk/repositories

# Install runtime dependencies
RUN apk add --no-cache tzdata curl bash busybox-extras

# Copy the binary from the builder stage
COPY --from=builder /app/app /app/app
RUN chmod +x /app/app

# Copy crontab and configuration templates (if any)
COPY crontab /etc/crontabs/root

# The app might need a config directory or messages.db if not mounted
# COPY config/ /app/config/

# Run crond in foreground with logging
CMD ["crond", "-f", "-l", "8"]
