#!/bin/bash
SOCKET="/data/data/com.termux/files/usr/tmp/juristic_gate.sock"
echo "--- RSJ-V3 JURISTIC VETO SYSTEM ---"
if [ ! -S "$SOCKET" ]; then echo "NO PENDING ACTIONS."; exit 1; fi
read -p "ENTER 6-DIGIT TOTP CODE: " code
echo -n "$code" | nc -U "$SOCKET"
