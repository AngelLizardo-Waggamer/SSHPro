package storage

import "fmt"

// Host represents a configured SSH host.
type Host struct {
	Name string `json:"name"`
	User string `json:"user"`
	IP   string `json:"ip"`
	Key  string `json:"key,omitempty"`
	Port int    `json:"port,omitempty"`
}

// PortOrDefault returns the configured port or 22 when empty.
func (h Host) PortOrDefault() int {
	if h.Port == 0 {
		return 22
	}
	return h.Port
}

// Address returns the user@host string.
func (h Host) Address() string {
	return fmt.Sprintf("%s@%s", h.User, h.IP)
}
