FROM golang:alpine AS builder
WORKDIR /app
RUN apk add --no-cache git protobuf protobuf-dev

COPY worker_grpc.go .
# Create proto dir just in case
RUN mkdir -p proto
# Create a dummy proto file if missing to prevent build error
RUN echo 'syntax="proto3"; package rsj; option go_package="./proto"; message EvidencePacket{string hash=1;string raw_data=2;string timestamp=3;} message WorkerResponse{string status=1;string message=2;int32 processing_time_ms=3;} service WorkerNode{rpc SubmitEvidence(EvidencePacket) returns(WorkerResponse);}' > proto/rsj.proto

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
