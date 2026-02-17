#!/bin/zsh

# rsj-v3-monorepo Deployment Script
# Managed by Principal DevOps Architect

set -e

# Ensure we are in the monorepo root
BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$BASE_DIR"

# Add Go binaries to PATH
export PATH=$PATH:$HOME/go/bin

echo "--- [RSJ-V3] MONOREPO DEPLOYMENT INITIATED ---"

# 1. Service Monitoring Integration
check_service_health() {
    local service=$1
    echo "[CHECK] Monitoring service: $service"
    if sv status "$service" 2>/dev/null | grep -q "^run:"; then
        echo "✅ Service $service is healthy."
    else
        echo "⚠️ Service $service is NOT running via 'sv'. Checking process..."
        if pgrep -x "$service" > /dev/null || pgrep -f "$service" > /dev/null; then
            echo "✅ Process found."
        else
            echo "❌ $service is DOWN."
            if [[ "$service" != "mariadbd" ]]; then return 1; fi
        fi
    fi
}

# 2. Database Connectivity
validate_db_connectivity() {
    echo "[CHECK] Validating Database Connectivity..."
    if mariadb-admin -u root ping > /dev/null 2>&1; then
        echo "✅ MariaDB is responsive."
    else
        echo "❌ MariaDB is UNREACHABLE."
        exit 1
    fi
}

# 3. Proto Generation
generate_protos() {
    echo "[BUILD] Generating gRPC bindings..."
    # Gatekeeper Protos
    protoc --proto_path=proto \
           --go_out=pkg/juristic --go_opt=paths=source_relative \
           --go-grpc_out=pkg/juristic --go-grpc_opt=paths=source_relative \
           proto/juristic.proto
           
    # Worker Protos
    mkdir -p pkg/worker
    protoc --proto_path=proto \
           --go_out=pkg/worker --go_opt=paths=source_relative \
           --go-grpc_out=pkg/worker --go-grpc_opt=paths=source_relative \
           proto/rsj.proto
           
    # Tidy worker pkg
    (cd pkg/worker && go mod tidy)
}

# 4. Monorepo Service Handling
deploy_services() {
    echo "[DEPLOY] Building Gatekeeper..."
    cd services/gatekeeper
    
    # We update main.go to use the correct import if it's not already correct
    sed -i 's|rsj-v3/juristic/juristic|rsj-v3-monorepo/pkg/juristic|g' main.go
    
    go build -o gatekeeper main.go
    
    if ! pgrep -x "gatekeeper" > /dev/null; then
        ./gatekeeper > gatekeeper.log 2>&1 &
        echo "🚀 Gatekeeper launched."
    else
        echo "ℹ️ Gatekeeper already running (Restarting...)"
        pkill -x "gatekeeper" || true
        ./gatekeeper > gatekeeper.log 2>&1 &
        echo "🚀 Gatekeeper restarted."
    fi
    cd ../..
}

# 5. Containerization (Cloud Worker)
build_cloud_worker() {
    echo "[DOCKER] Building Cloud Worker Image..."
    
    # Check for container runtime (Local)
    local runtime="docker"
    if command -v docker &> /dev/null; then
        echo "Using local runtime: docker"
        $runtime build -t rsj-v3-cloud-worker -f services/cloud-worker/Dockerfile . || {
            echo "⚠️ Local build failed. Trying remote build..."
            trigger_remote_build
        }
    elif command -v podman &> /dev/null; then
        echo "Using local runtime: podman"
        podman build -t rsj-v3-cloud-worker -f services/cloud-worker/Dockerfile . || {
            echo "⚠️ Local build failed. Trying remote build..."
            trigger_remote_build
        }
    else
        echo "⚠️ No local container runtime found. defaulting to Remote Build..."
        trigger_remote_build
    fi
}

trigger_remote_build() {
    if command -v gcloud &> /dev/null; then
        echo "[CLOUD] Submitting build to Google Cloud Build..."
        gcloud builds submit --config=cloudbuild.yaml .
    else
        echo "❌ neither Docker nor gcloud found. Cannot build container."
        return 1
    fi
}

# Execution
validate_db_connectivity
check_service_health "mariadbd"
generate_protos
deploy_services
build_cloud_worker

echo "--- [RSJ-V3] DEPLOYMENT COMPLETE ---"
