# Managed by Principal DevOps Architect

# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git protobuf protobuf-dev

# Copy dependency files
# Note: In a monorepo, we often copy the root go.work or individual go.mods
# For simplicity, we'll assume we build from the service directory but have access to shared libs if context allows
# However, standard Docker builds don't allow accessing parent dirs unless context is root.

# Strategy: Build from Monorepo Root
# command: docker build -f services/cloud-worker/Dockerfile .

# 1. Copy Shared Protos/Libs
COPY proto/ ./proto/
COPY pkg/ ./pkg/

# 2. Copy Service Source
COPY services/cloud-worker/ .

# 3. Initialize/Tidy Module (if not pre-baked)
# If go.mod exists in service, use it. If not, init.
RUN [ -f go.mod ] || go mod init rsj-v3-monorepo/services/cloud-worker

# IMPORTANT: Link to local shared package
RUN go mod edit -replace rsj-v3-monorepo/pkg/worker=./pkg/worker || true

RUN go mod tidy

# 4. Build
# We use the shared pkg now, so no need to regenerate protos here if we copied generated code
# Target ONLY worker_grpc.go to avoid conflict with legacy worker.go
RUN CGO_ENABLED=0 GOOS=linux go build -o cloud-worker worker_grpc.go

# Production Stage
FROM alpine:latest

WORKDIR /root/

# Install CA Certs
RUN apk --no-cache add ca-certificates

# Copy Binary
COPY --from=builder /app/cloud-worker .

# Copy Config (Optional, usually mounted)
# COPY .env . 

EXPOSE 8081

CMD ["./cloud-worker"]
