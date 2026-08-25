#!/bin/sh
# entrypoint.sh — AmneziaWG Mikrotik Panel
# Compatible with Alpine sh (no bash required) and MikroTik RouterOS v7 Container

# Ensure persistent data directory exists (mapped via /container/mounts in RouterOS)
mkdir -p /app/data

echo "[entrypoint] Waiting for /dev/net/tun to become available..."
# MikroTik RouterOS exposes /dev/net/tun for userspace TUN apps like amneziawg-go
RETRIES=10
while [ ! -c /dev/net/tun ] && [ "$RETRIES" -gt 0 ]; do
    sleep 1
    RETRIES=$((RETRIES - 1))
done

if [ ! -c /dev/net/tun ]; then
    echo "[entrypoint] WARNING: /dev/net/tun not found. amneziawg-go may fail."
    echo "[entrypoint] Ensure the MikroTik container package supports TUN devices on your hardware."
fi

echo "[entrypoint] Initializing awg0 interface via amneziawg-go (userspace)..."
# Run amneziawg-go in background; it creates the awg0 TUN interface and exits.
# Do NOT use set -e here — failure should be non-fatal so the web panel still starts.
if ! ip link show awg0 > /dev/null 2>&1; then
    amneziawg-go awg0 || echo "[entrypoint] WARNING: amneziawg-go failed to create awg0. Panel will start but VPN may not work."
    echo "[entrypoint] Interface awg0 created via amneziawg-go (userspace TUN)."
else
    echo "[entrypoint] Interface awg0 already exists, skipping creation."
fi

echo "[entrypoint] Starting Amnezia Mikrotik Panel web server..."
exec /app/panel
