# Build stage
FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application for target platform
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o webshell main.go

# Final stage
FROM --platform=$TARGETPLATFORM alpine:latest

# Install ca-certificates for HTTPS support
# Use --no-cache and update index first to avoid QEMU issues
RUN apk update && apk --no-cache add ca-certificates tzdata && rm -rf /var/cache/apk/*

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/webshell .
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./webshell"]
