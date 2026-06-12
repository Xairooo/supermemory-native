# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build clean static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o supermemory-native ./cmd/supermemory-native

# Run stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates in case they need to connect to cloud embedding APIs (like Gemini)
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/supermemory-native /app/supermemory-native

# Default port
EXPOSE 6767

# Run binary
ENTRYPOINT ["/app/supermemory-native"]
