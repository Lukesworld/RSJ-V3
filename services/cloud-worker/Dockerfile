FROM golang:alpine AS builder
WORKDIR /app
RUN apk add --no-cache git protobuf protobuf-dev

COPY worker_grpc.go .
COPY proto/ ./proto/

# Initialize Module
RUN go mod init rsj-worker
RUN go mod tidy

# Generate Code
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/rsj.proto

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o rsj-worker worker_grpc.go

# Production Image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/rsj-worker .
EXPOSE 50051
CMD ["./rsj-worker"]
