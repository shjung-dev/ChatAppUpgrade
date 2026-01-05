# Stage 1: Build
FROM golang:1.25.5-alpine AS builder

# Install git & certificates for go modules
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the Go binary
RUN go build -o main .

# Stage 2: Minimal runtime image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
