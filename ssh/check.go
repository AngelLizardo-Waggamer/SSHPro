package ssh

import (
	"net"
	"strconv"
	"sync"
	"time"

	"ssh-pro/storage"
)

// HostStatus holds the connectivity status and latency for an SSH host.
type HostStatus struct {
	Host    storage.Host
	Online  bool
	Latency time.Duration
	Error   string
}

// CheckHost verifies TCP connectivity to the SSH port of the given host.
func CheckHost(host storage.Host, timeout time.Duration) HostStatus {
	addr := net.JoinHostPort(host.IP, strconv.Itoa(host.PortOrDefault()))
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return HostStatus{
			Host:   host,
			Online: false,
			Error:  err.Error(),
		}
	}
	_ = conn.Close()
	return HostStatus{
		Host:    host,
		Online:  true,
		Latency: time.Since(start),
	}
}

// CheckHostsConcurrently checks connectivity to multiple SSH hosts concurrently.
func CheckHostsConcurrently(hosts []storage.Host, timeout time.Duration) []HostStatus {
	results := make([]HostStatus, len(hosts))
	var wg sync.WaitGroup

	for i, h := range hosts {
		wg.Add(1)
		go func(idx int, target storage.Host) {
			defer wg.Done()
			results[idx] = CheckHost(target, timeout)
		}(i, h)
	}

	wg.Wait()
	return results
}
