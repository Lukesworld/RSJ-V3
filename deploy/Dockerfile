FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install git/protobuf just in case
RUN apk add --no-cache git protobuf

# Copy the pre-calculated dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o rsj-worker worker_grpc.go

# Production Stage
FROM alpine:latest
WORKDIR /root/
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/rsj-worker .
EXPOSE 50051
CMD ["./rsj-worker"]
