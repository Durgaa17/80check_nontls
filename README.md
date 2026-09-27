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

### Build

```bash
go build -o 80check 80check.go
```

### Run

```bash
./80check 8.8.8.8
./80check 192.168.1.0/24
./80check -t 2s -w 20 10.0.0.0/20
```

| Flag | Description | Default |
|------|-------------|---------|
| `-t` | Timeout (e.g. `3s`, `1500ms`) | 3s |
| `-w` | Max concurrent workers | 40 |

### Cross-compile (optional)

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o 80check-linux 80check.go

# Windows
GOOS=windows GOARCH=amd64 go build -o 80check.exe 80check.go

# Android (Termux can run the binary if built for arm64)
GOOS=linux GOARCH=arm64 go build -o 80check-arm64 80check.go
```

---

## Example output

```
==================================================
  80check_nontls  (port-80 scanner)
==================================================
[*] Target      : 192.168.1.0/24
[*] Hosts       : 254
[*] Port        : 80
[*] Timeout     : 3.0s
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
