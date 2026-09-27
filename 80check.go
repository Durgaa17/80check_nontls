package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout = 3 * time.Second
	port           = 80
	maxWorkers     = 40
	maxHosts       = 1024
)

type result struct {
	ip     string
	status string // reachable | timeout | error
	ms     int64
	line   string
}

func probe(ip string, timeout time.Duration) result {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, strconv.Itoa(port)), timeout)
	ms := time.Since(start).Milliseconds()

	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return result{ip: ip, status: "timeout", ms: ms, line: "No response (timeout)"}
		}
		return result{ip: ip, status: "error", ms: ms, line: err.Error()}
	}
	defer conn.Close()

	// Try to read a simple HTTP status line
	_ = conn.SetDeadline(time.Now().Add(timeout))
	_, _ = conn.Write([]byte("GET / HTTP/1.0\r\nHost: " + ip + "\r\n\r\n"))
	buf := make([]byte, 128)
	n, _ := conn.Read(buf)
	line := "Port 80 open (no status line)"
	if n > 0 {
		first := strings.SplitN(string(buf[:n]), "\r\n", 2)[0]
		first = strings.TrimSpace(first)
		if first != "" {
			line = first
		}
	}
	return result{ip: ip, status: "reachable", ms: ms, line: line}
}

// parseCIDR returns list of host IPs (IPv4 only for simplicity)
func parseCIDR(target string) ([]string, error) {
	// Single IP?
	if ip := net.ParseIP(target); ip != nil {
		if ip.To4() != nil {
			return []string{ip.String()}, nil
		}
		return nil, fmt.Errorf("IPv6 not supported in this simple version")
	}

	// CIDR
	_, network, err := net.ParseCIDR(target)
	if err != nil {
		return nil, fmt.Errorf("invalid IP or CIDR: %s", target)
	}

	// Only IPv4
	if network.IP.To4() == nil {
		return nil, fmt.Errorf("IPv6 CIDR not supported")
	}

	var hosts []string
	for ip := network.IP.Mask(network.Mask); network.Contains(ip); inc(ip) {
		hosts = append(hosts, ip.String())
	}

	// remove network & broadcast if more than 2 hosts
	if len(hosts) > 2 {
		hosts = hosts[1 : len(hosts)-1]
	}
	return hosts, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func main() {
	timeoutFlag := flag.Duration("t", defaultTimeout, "Timeout duration (e.g. 3s)")
	workersFlag := flag.Int("w", maxWorkers, "Max concurrent workers")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "80check_nontls - port-80 (non-TLS) reachability scanner\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [flags] <IP or CIDR>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Examples:\n  %s 8.8.8.8\n  %s 192.168.1.0/24\n  %s -t 2s -w 20 10.0.0.0/20\n\n", os.Args[0], os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	target := flag.Arg(0)
	if target == "" {
		fmt.Print("Enter IP or CIDR (e.g. 192.168.1.1 or 192.168.1.0/24): ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			target = strings.TrimSpace(scanner.Text())
		}
		if target == "" {
			fmt.Println("No target given.")
			os.Exit(1)
		}
	}

	hosts, err := parseCIDR(target)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(hosts) > maxHosts {
		fmt.Printf("Range too big (%d hosts). Max %d allowed.\n", len(hosts), maxHosts)
		os.Exit(1)
	}

	fmt.Println("==================================================")
	fmt.Println("  80check_nontls  (port-80 scanner)  [Go]")
	fmt.Println("==================================================")
	fmt.Printf("[*] Target      : %s\n", target)
	fmt.Printf("[*] Hosts       : %d\n", len(hosts))
	fmt.Printf("[*] Port        : %d\n", port)
	fmt.Printf("[*] Timeout     : %v\n", *timeoutFlag)
	fmt.Printf("[*] Started     : %s\n", time.Now().Format("15:04:05"))
	fmt.Println("--------------------------------------------------")

	results := make([]result, 0, len(hosts))
	var mu sync.Mutex
	sem := make(chan struct{}, *workersFlag)
	var wg sync.WaitGroup

	for _, ip := range hosts {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			r := probe(ip, *timeoutFlag)
			mu.Lock()
			results = append(results, r)
			mu.Unlock()

			switch r.status {
			case "reachable":
				fmt.Printf("[+] %s:80   %dms   REACHABLE   %s\n", r.ip, r.ms, r.line)
			case "timeout":
				fmt.Printf("[-] %s:80   %dms   NO RESPONSE\n", r.ip, r.ms)
			default:
				fmt.Printf("[!] %s:80   %dms   NOT REACHABLE\n", r.ip, r.ms)
			}
		}(ip)
	}
	wg.Wait()

	// sort results by IP
	sort.Slice(results, func(i, j int) bool {
		a := net.ParseIP(results[i].ip).To4()
		b := net.ParseIP(results[j].ip).To4()
		if a == nil || b == nil {
			return results[i].ip < results[j].ip
		}
		return uint32(a[0])<<24|uint32(a[1])<<16|uint32(a[2])<<8|uint32(a[3]) <
			uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3])
	})

	reachable := 0
	for _, r := range results {
		if r.status == "reachable" {
			reachable++
		}
	}

	fmt.Println("--------------------------------------------------")
	fmt.Printf("Finished.  Reachable: %d / %d\n", reachable, len(results))
	fmt.Println("==================================================")

	if reachable > 0 {
		fmt.Println("\nREACHABLE hosts:")
		for _, r := range results {
			if r.status == "reachable" {
				fmt.Printf("  %s:80  →  %dms  |  %s\n", r.ip, r.ms, r.line)
			}
		}
	}
}
