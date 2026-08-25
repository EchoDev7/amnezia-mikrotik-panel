# 📋 PROJECT MANIFESTO & SYSTEM INSTRUCTIONS (AG_INSTRUCTIONS.md)

## 1. Project Goal & Vision
- **Objective:** Build a ultra-lightweight, high-performance, single-binary Web Panel in Go (Golang) to manage AmneziaWG (v3.1) peers.
- **Target Platform:** Docker Container running natively on MikroTik RouterOS v7.
- **Core Constraints:** RAM usage under 20MB, zero runtime node.js dependencies, 100% data persistence across server reboots, and Zero-Config deployment.

## 2. Tech Stack & References
- **Backend Language:** Go (Golang 1.22+)
- **VPN Core:** AmneziaWG Core & Tools v3.1 (`amneziawg-go`, `amneziawg-tools`)
- **Database:** SQLite (Embedded, CGO Disabled / Pure-Go driver)
- **Frontend:** Go HTML Templates + Embedded Tailwind CSS
- **Base OS:** Alpine Linux 3.20+ (Multi-Stage Build)

## 3. AmneziaWG 3.1 Protocol Requirements
The panel MUST support all new v3.1 anti-DPI obfuscation parameters in client configs:
- **Junk Packets:** `Jc`, `Jmin`, `Jmax`
- **Padding Controls:** `S1`, `S2`, `S3`, `S4`
- **Magic Headers:** `H1`, `H2`, `H3`, `H4`
- **Dual-Stack:** Simultaneous IPv4 & IPv6 allocation per peer.

## 4. Business Logic & Persistence Rules (CRITICAL)

1. **User Status Lifecycle (Never Delete Expired Users):**
   - User statuses: `active`, `expired`, `limited`, `disabled`.
   - When a user hits volume limit or expiration date: DO NOT delete the record from SQLite.
   - Execute `awg set awg0 peer <PUBKEY> remove` to disconnect active connection immediately.
   - Update user status in database to `expired` or `limited`.
   - On container startup/restart: ONLY peers with `active` status must be injected into the `awg0` interface.

2. **Delta Bandwidth Accounting (Reboot-Proof):**
   - Bandwidth counters from `awg show awg0 dump` reset to 0 on server reboot.
   - Must use Delta formula: `Total_Bytes = Historic_Bytes + (Current_RxTx - Session_Start_RxTx)`.
   - Background worker (every 30s) updates and commits cumulative bytes to SQLite `/app/data/panel.db`.

## 5. Repository Structure Standard
```text
/
├── Dockerfile             # Multi-stage, Multi-arch build
├── entrypoint.sh          # Interface startup script
├── go.mod / go.sum
├── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── models/
│   ├── service/           # AWG CLI execution wrapper
│   ├── worker/            # Cron worker (Delta usage & Expiration)
│   └── web/               # API & Web HTTP handlers
├── templates/             # HTML Templates
└── .github/
    └── workflows/
        └── docker.yml     # Multi-Arch Docker Hub Push