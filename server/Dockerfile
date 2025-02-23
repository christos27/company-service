# Start with a lightweight Go image
FROM golang:1.24 AS builder

# Set environment variables
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the Go binary
RUN go build -o server .

# Use a lightweight image for the final container
FROM alpine:latest

# Set working directory in the container
WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/server .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./server"]
