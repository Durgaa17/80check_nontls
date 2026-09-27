# 80check_nontls

Simple **port-80 (non-TLS)** reachability scanner.

Checks whether hosts answer on **TCP port 80** and optionally reads the HTTP status line.  
Supports **single IP** or **CIDR** (e.g. `192.168.1.0/24`).

Works on:

- Windows CMD / PowerShell
- Linux
- Termux (Android)
- Pydroid3
- Any system with Python 3 or Go

Inspired by the [sivdip](https://github.com/Jeeva-zone/sivdip) web port-80 scanner.

---

## Features

- Single IP or CIDR range
- Concurrent scanning (configurable workers)
- Reports: **REACHABLE** / **NO RESPONSE** / **NOT REACHABLE**
- Tries to read the HTTP status line when possible
- Safety limit (max 1024 hosts) so large ranges don’t freeze mobile devices
- Pure Python version (no external dependencies)
- Fast Go version

---

## Python version

### Requirements

- Python 3.6+ (standard library only)

### Run

```bash
# Interactive
python 80check.py

# Single IP
python 80check.py 8.8.8.8

# CIDR
python 80check.py 192.168.1.0/24
python 80check.py 10.0.0.0/20

# Options
python 80check.py -t 2 -w 20 192.168.1.0/24
```

| Flag | Description | Default |
|------|-------------|---------|
| `-t`, `--timeout` | Timeout in seconds | 3.0 |
| `-w`, `--workers` | Max concurrent workers | 40 |

### Termux / Pydroid3

```bash
# Termux
pkg install python
python 80check.py 192.168.1.0/24

# Pydroid3 – just open the file and run
```

---

## Go version

Fast concurrent implementation written in pure Go (no external dependencies).

### Requirements

- Go 1.18 or newer (any recent version works)

Install Go:

```bash
# Linux / macOS / WSL
# See https://go.dev/dl/

# Termux (Android)
pkg install golang
```

### Build

```bash
# Simple local build
go build -o 80check 80check.go

# Smaller binary (strip debug info)
go build -ldflags="-s -w" -o 80check 80check.go
```

After building you get a single static binary named `80check` (or `80check.exe` on Windows).

### Run

```bash
# Interactive (asks for target)
./80check

# Single IP
./80check 8.8.8.8
./80check 1.1.1.1

# CIDR ranges
./80check 192.168.1.0/24
./80check 10.0.0.0/20
./80check 172.16.0.0/16

# With flags
./80check -t 2s -w 30 192.168.1.0/24
./80check -t 1500ms -w 50 10.0.0.0/22
```

### Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-t` | Connection timeout | `3s` | `-t 2s`, `-t 1500ms`, `-t 1s` |
| `-w` | Max concurrent workers | `40` | `-w 20`, `-w 100` |
| `-h` | Show help | – | `./80check -h` |

### Cross-compile (build for other platforms)

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o 80check-linux 80check.go

# Windows amd64
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o 80check.exe 80check.go

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o 80check-macos 80check.go

# Android / Termux (arm64)
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o 80check-arm64 80check.go

# Android 32-bit (older devices)
GOOS=linux GOARCH=arm go build -ldflags="-s -w" -o 80check-arm 80check.go
```

### Termux (Android)

```bash
# Install Go
pkg update && pkg install golang

# Clone or download the files
git clone https://github.com/Durgaa17/80check_nontls.git
cd 80check_nontls

# Build
go build -ldflags="-s -w" -o 80check 80check.go

# Run
./80check 192.168.1.0/24
./80check -t 2s -w 20 10.0.0.0/24
```

You can also download a pre-built `80check-arm64` binary and run it directly after `chmod +x`.

### Performance tips

- Default 40 workers is a good balance for most networks and phones.
- On a powerful machine / fast LAN you can raise workers: `-w 100` or higher.
- On slow mobile data or weak devices lower it: `-w 10` or `-w 20`.
- Timeout of `1s`–`2s` is usually enough on LAN; keep `3s` for the internet.
- The tool automatically skips network & broadcast addresses and limits to 1024 hosts for safety.

### Notes

- Only **IPv4** is supported in the current Go version (same as the Python version for simplicity).
- The scanner connects to **TCP port 80** and tries to read the first HTTP response line.
- Any HTTP response (even 404 / 403 / 302) counts as **REACHABLE**.
- Connection refused → **NOT REACHABLE**
- Timeout → **NO RESPONSE**

---

## Example output

```
==================================================
  80check_nontls  (port-80 scanner)  [Go]
==================================================
[*] Target      : 192.168.1.0/24
[*] Hosts       : 254
[*] Port        : 80
[*] Timeout     : 3s
[*] Started     : 14:22:01
--------------------------------------------------
[+] 192.168.1.1:80   12ms   REACHABLE   HTTP/1.1 200 OK
[+] 192.168.1.10:80  45ms   REACHABLE   HTTP/1.0 302 Found
[-] 192.168.1.50:80  3001ms  NO RESPONSE
[!] 192.168.1.99:80  3ms    NOT REACHABLE
...
--------------------------------------------------
Finished.  Reachable: 3 / 254
==================================================

REACHABLE hosts:
  192.168.1.1:80  →  12ms  |  HTTP/1.1 200 OK
  192.168.1.10:80 →  45ms  |  HTTP/1.0 302 Found
```

---

## License

MIT
