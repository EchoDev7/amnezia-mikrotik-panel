#!/bin/sh
set -e

# Ensure data directory exists
mkdir -p /app/data

echo "Initializing awg0 interface..."
if ! ip link show awg0 > /dev/null 2>&1; then
    # Create the tun interface using amneziawg-go
    amneziawg-go awg0
    echo "Interface awg0 created via amneziawg-go."
else
    echo "Interface awg0 already exists."
fi

# Note: The panel application is responsible for configuring IP, 
# port, and peers on the awg0 interface at startup.

echo "Starting Amnezia Mikrotik Panel..."
exec /app/panel
