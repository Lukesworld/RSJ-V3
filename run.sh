#!/data/data/com.termux/files/usr/bin/bash

# Newman Ventures - RSJ-V3 Master Startup Script
# Environment: Termux / Android Sovereign Node

# 1. Load Environment
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
if [ -f "$DIR/.env" ]; then
    export $(grep -v '^#' "$DIR/.env" | sed 's/export //g' | xargs)
fi

LOG_DIR="$DIR/logs"
mkdir -p "$LOG_DIR"

echo "--- Initializing RSJ-V3 Sovereign Node ---"

# 2. Start Juristic Gatekeeper (Control Plane)
echo "[1/3] Starting Gatekeeper..."
./bin/gatekeeper > "$LOG_DIR/gatekeeper.log" 2>&1 &
GATEKEEPER_PID=$!

# 3. Start Cloud Worker (AI & Evidence Processor)
echo "[2/3] Starting Cloud Worker..."
./bin/cloud-worker > "$LOG_DIR/cloud-worker.log" 2>&1 &
WORKER_PID=$!

# 4. Start System Guardian (Integrity & Authentication)
# This is the master process that stays in foreground
echo "[3/3] Starting System Guardian..."
echo "------------------------------------------"
./bin/rsj-guardian

# Cleanup on exit
trap "kill $GATEKEEPER_PID $WORKER_PID; exit" SIGINT SIGTERM
