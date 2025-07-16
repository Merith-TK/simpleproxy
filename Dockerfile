# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY ./cmd/simpleproxy ./cmd/simpleproxy

# Build the binary with optimizations
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/simpleproxy ./cmd/simpleproxy

# Final Minimal Image
FROM alpine:latest

WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/simpleproxy .

# Run the binary
ENTRYPOINT ["./simpleproxy"]
