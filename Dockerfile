FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev linux-headers bash

WORKDIR /build

# Build amneziawg-go (Userspace Go implementation)
RUN git clone https://github.com/amnezia-vpn/amneziawg-go.git && \
    cd amneziawg-go && \
    make && \
    make install

# Build amneziawg-tools (awg, awg-quick)
RUN git clone https://github.com/amnezia-vpn/amneziawg-tools.git && \
    cd amneziawg-tools/src && \
    make && \
    make install

# Build Web Panel
WORKDIR /app

# Copy go.mod first
COPY go.mod ./

# Copy full source code so go mod tidy can inspect imports
COPY . .

# Tidy and download all dependencies inside the container to ensure go.sum is synced
RUN go mod tidy && go mod download

# Compile Go app with CGO disabled (pure Go SQLite driver)
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o panel .

# Final Stage
FROM alpine:3.20

# Install runtime dependencies (iptables for awg-quick/routing, iproute2 for ip commands)
RUN apk add --no-cache bash iptables iproute2

# Copy compiled binaries from builder
COPY --from=builder /usr/bin/amneziawg-go /usr/bin/amneziawg-go
COPY --from=builder /usr/bin/awg /usr/bin/awg
COPY --from=builder /usr/bin/awg-quick /usr/bin/awg-quick
COPY --from=builder /app/panel /app/panel

# Copy HTML templates (required at runtime by routes.go)
COPY --from=builder /app/templates /app/templates

# Copy entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Create persistent data directory
RUN mkdir -p /app/data

WORKDIR /app
EXPOSE 8080 51820/udp

ENTRYPOINT ["/app/entrypoint.sh"]
