#!/usr/bin/env python3
"""
80check_nontls - Simple port-80 (non-TLS) reachability scanner
Works on: Windows CMD, Linux, Termux, Pydroid3
Supports single IP or CIDR (e.g. 192.168.1.0/24)
"""

import socket
import ipaddress
import concurrent.futures
import sys
import argparse
from datetime import datetime

TIMEOUT = 3.0
PORT = 80
MAX_WORKERS = 40
MAX_HOSTS = 1024


def probe(ip: str):
    """Probe one IP on port 80. Returns (status, ms, line)"""
    start = datetime.now()
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(TIMEOUT)
            result = s.connect_ex((str(ip), PORT))
            ms = int((datetime.now() - start).total_seconds() * 1000)

            if result == 0:
                try:
                    s.sendall(b"GET / HTTP/1.0\r\nHost: " + str(ip).encode() + b"\r\n\r\n")
                    data = s.recv(128).decode(errors="ignore")
                    first_line = data.split("\r\n")[0].strip() if data else "HTTP response"
                    return "reachable", ms, first_line
                except Exception:
                    return "reachable", ms, "Port 80 open (no status line)"
            else:
                return "error", ms, "Connection refused / closed"
    except socket.timeout:
        ms = int((datetime.now() - start).total_seconds() * 1000)
        return "timeout", ms, "No response (timeout)"
    except Exception as e:
        ms = int((datetime.now() - start).total_seconds() * 1000)
        return "error", ms, str(e)


def main():
    parser = argparse.ArgumentParser(
        description="port-80 (non-TLS) reachability scanner. Single IP or CIDR.",
        epilog="Examples:\n  python 80check.py 8.8.8.8\n  python 80check.py 192.168.1.0/24\n  python 80check.py 10.0.0.0/20",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("target", nargs="?", help="IP address or CIDR (e.g. 192.168.1.1 or 192.168.1.0/24)")
    parser.add_argument("-t", "--timeout", type=float, default=TIMEOUT, help=f"Timeout in seconds (default: {TIMEOUT})")
    parser.add_argument("-w", "--workers", type=int, default=MAX_WORKERS, help=f"Max concurrent workers (default: {MAX_WORKERS})")
    args = parser.parse_args()

    target = args.target
    if not target:
        target = input("Enter IP or CIDR (e.g. 192.168.1.1 or 192.168.1.0/24): ").strip()
        if not target:
            print("No target given.")
            sys.exit(1)

    global TIMEOUT
    TIMEOUT = args.timeout

    try:
        network = ipaddress.ip_network(target, strict=False)
    except ValueError:
        print(f"Invalid IP / CIDR: {target}")
        sys.exit(1)

    hosts = list(network.hosts()) if network.num_addresses > 1 else [network.network_address]

    if len(hosts) > MAX_HOSTS:
        print(f"Range too big ({len(hosts)} hosts). Max {MAX_HOSTS} allowed.")
        sys.exit(1)

    print("=" * 50)
    print("  80check_nontls  (port-80 scanner)")
    print("=" * 50)
    print(f"[*] Target      : {target}")
    print(f"[*] Hosts       : {len(hosts)}")
    print(f"[*] Port        : {PORT}")
    print(f"[*] Timeout     : {TIMEOUT}s")
    print(f"[*] Started     : {datetime.now().strftime('%H:%M:%S')}")
    print("-" * 50)

    results = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=args.workers) as executor:
        future_map = {executor.submit(probe, str(ip)): ip for ip in hosts}
        for future in concurrent.futures.as_completed(future_map):
            ip = future_map[future]
            try:
                status, ms, line = future.result()
                results.append((ip, status, ms, line))
                if status == "reachable":
                    print(f"[+] {ip}:80   {ms}ms   REACHABLE   {line}")
                elif status == "timeout":
                    print(f"[-] {ip}:80   {ms}ms   NO RESPONSE")
                else:
                    print(f"[!] {ip}:80   {ms}ms   NOT REACHABLE")
            except Exception as e:
                print(f"[!] {ip}:80   ERROR  {e}")

    reachable = [r for r in results if r[1] == "reachable"]
    print("-" * 50)
    print(f"Finished.  Reachable: {len(reachable)} / {len(results)}")
    print("=" * 50)

    if reachable:
        print("\nREACHABLE hosts:")
        for ip, status, ms, line in sorted(reachable, key=lambda x: int(ipaddress.ip_address(str(x[0])))):
            print(f"  {ip}:80  →  {ms}ms  |  {line}")


if __name__ == "__main__":
    main()
