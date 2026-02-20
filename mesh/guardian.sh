#!/bin/bash
# Load configuration from project root
if [ -f "../../.env" ]; then
    source "../../.env"
    echo "loaded .env: JURISTIC_PORT=$JURISTIC_PORT"
else
    echo "Warning: .env not found at ../../.env"
    export JURISTIC_PORT=50051 # Default fallback
fi

BINARY="rsj-edge"
LAUNCHER="uplink.sh"
while true; do
    if [ ! -f "$LAUNCHER" ]; then
        echo "🛠️ Creating Launcher..."
        cat > "$LAUNCHER" << INNEREOF
#!/bin/bash
mkdir -p /etc/ssl/certs
[ -f "/usr/etc/tls/cert.pem" ] && cat /usr/etc/tls/cert.pem > /etc/ssl/certs/ca-certificates.crt
export JURISTIC_PORT=$JURISTIC_PORT
export AGENT_PORT=$AGENT_PORT
./rsj-edge
INNEREOF
        chmod +x "$LAUNCHER"
    fi
    # Re-create launcher on every run to ensure env vars are fresh?
    # Or just export them here and hope termux-chroot passes them.
    # The cat << INNEREOF expands variables if not quoted 'INNEREOF'.
    # I changed 'INNEREOF' to INNEREOF to allow expansion.
    
    termux-chroot ./$LAUNCHER
    sleep 15
done
