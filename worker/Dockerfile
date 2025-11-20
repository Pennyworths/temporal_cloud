# ---------- Build stage ----------
    FROM --platform=linux/amd64 golang:1.23-alpine AS builder

    # Install necessary tools (optional, usually not needed)
    # RUN apk add --no-cache ca-certificates
    
    WORKDIR /app
    
    # Copy go module files and download deps
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the rest of the source code
    COPY . .
    
    # Build a static Linux binary
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o worker .
    
    # ---------- Runtime stage ----------
    FROM --platform=linux/amd64 alpine:3.20
    
    WORKDIR /app
    
    # Copy compiled binary from builder stage
    COPY --from=builder /app/worker .
    
    # If you need CA certs for TLS (Temporal Cloud), add the line below:
    # RUN apk add --no-cache ca-certificates
    
    # Entry point: start the Temporal worker
    ENTRYPOINT ["/app/worker"]
    