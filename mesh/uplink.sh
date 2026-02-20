#!/bin/bash
mkdir -p /etc/ssl/certs
[ -f "/usr/etc/tls/cert.pem" ] && cat /usr/etc/tls/cert.pem > /etc/ssl/certs/ca-certificates.crt
./rsj-edge
