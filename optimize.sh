#!/bin/bash
# Advanced Optimization Script for Local Services (Termux/Linux)

SERVICE_NAME="gatekeeper"
CORE_MASK="03" # Bind to cores 0 and 1 (binary 0011 -> hex 03)

echo "--- [RSJ-V3] ADVANCED OPTIMIZATION PROTOCOL ---"

# 1. Process Pinning (Taskset)
echo "[OPTIMIZE] Attempting to pin '$SERVICE_NAME' to Cores 0-1..."
PID=$(pgrep -x "$SERVICE_NAME")

if [ -n "$PID" ]; then
    if command -v taskset &> /dev/null; then
        taskset -p $CORE_MASK $PID
        if [ $? -eq 0 ]; then
            echo "✅ SUCCESS: $SERVICE_NAME (PID $PID) bound to mask $CORE_MASK."
        else
            echo "⚠️ FAILED: Could not set affinity. (Requires root/privilege?)"
        fi
    else
        echo "❌ SKIP: 'taskset' not found."
    fi
else
    echo "❌ SKIP: Service '$SERVICE_NAME' not running."
fi

# 2. Ramdisk (Tmpfs)
# On Cloud Run, /tmp is already in-memory.
# On Local, we check if we can mount.
echo "[OPTIMIZE] Checking Ramdisk Status..."
if mount | grep -q "tmpfs on /tmp"; then
    echo "✅ /tmp is already a tmpfs (Ramdisk)."
else
    echo "ℹ️ /tmp is physically backed. Recommendation: 'mount -t tmpfs tmpfs /tmp' (Requires Root)."
fi

echo "--- OPTIMIZATION COMPLETE ---"
