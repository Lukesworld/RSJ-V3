#!/bin/bash

echo "[RSJ-V3] Compiling Guardian (Memory Optimization)..."
go build -o rsj-guardian-bin main.go

echo "[RSJ-V3] Watchdog Active. Protecting Process..."
while true; do
    ./rsj-guardian-bin
    echo "[RSJ-V3] Process interrupted! Re-engaging Sovereign Node in 3 seconds..."
    sleep 3
done
