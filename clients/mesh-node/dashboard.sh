#!/bin/bash
LOG_FILE="nexus.log"
while true; do
    clear
    echo "===================================================="
    echo "        RSJ-V3 NEURAL LINK DASHBOARD               "
    echo "===================================================="
    echo " Current Time: $(date)"
    echo "----------------------------------------------------"
    if [ -f "$LOG_FILE" ]; then
        TOTAL=$(grep -c "ACK: secured" "$LOG_FILE")
        FAILURES=$(grep -c "Fail" "$LOG_FILE")
        LAST_ACK=$(grep "ACK: secured" "$LOG_FILE" | tail -1 | awk -F'ACK: secured | ' '{print $2}')
        echo " [+] TOTAL SUCCESSFUL UPLINKS: $TOTAL"
        echo " [!] TOTAL CONNECTION FAILS:   $FAILURES"
        echo " [>] LATEST STATUS:            $LAST_ACK"
        echo "----------------------------------------------------"
        echo " RECENT ACTIVITY (Last 5 Pulses):"
        tail -n 5 "$LOG_FILE"
    else
        echo " [!] No log data found yet. Waiting for first pulse..."
    fi
    echo "----------------------------------------------------"
    echo " (Ctrl+C to exit dashboard - Guardian remains active)"
    sleep 5
done
