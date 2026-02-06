# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scoundrel-server .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/scoundrel-server .

# Generate SSH host key
RUN mkdir -p .ssh && \
    apk add --no-cache openssh-keygen && \
    ssh-keygen -t ed25519 -f .ssh/id_ed25519 -N "" -C "scoundrel-server" && \
    apk del openssh-keygen

# Expose SSH port
EXPOSE 22

# Run the server
CMD ["./scoundrel-server"]
