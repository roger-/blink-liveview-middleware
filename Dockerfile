# Build stage
FROM golang:alpine AS builder

# Install ca-certificates and git
RUN apk add --no-cache ca-certificates git

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o blink-liveview-middleware main.go

# Runtime stage
FROM alpine:latest

# Install ffmpeg
RUN apk add --no-cache ffmpeg

# Copy the binary from builder
COPY --from=builder /build/blink-liveview-middleware /usr/local/bin/blink-liveview-middleware

# Set the entrypoint and default command
ENTRYPOINT ["blink-liveview-middleware"]
CMD ["liveview"]
