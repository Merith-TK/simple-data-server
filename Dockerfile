# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o simple-data-server .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/simple-data-server .

# Create data directory
RUN mkdir -p /data

# Set environment variables
ENV PORT=8080
ENV DATA_DIR=/data

EXPOSE 8080

CMD ["./simple-data-server"]
