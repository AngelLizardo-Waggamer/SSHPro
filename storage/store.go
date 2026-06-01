package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Store handles persistence for SSH hosts.
type Store struct {
	Path string
}

// DefaultPath returns the default JSON file path under the user's home.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".sshpro", "configured_hosts.json"), nil
}

// NewStore creates a new Store for the given path.
func NewStore(path string) *Store {
	return &Store{Path: path}
}

// Ensure creates the storage file if it doesn't exist.
func (s *Store) Ensure() error {
	if s.Path == "" {
		return fmt.Errorf("storage path is empty")
	}
	if _, err := os.Stat(s.Path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat storage file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return fmt.Errorf("create storage dir: %w", err)
	}
	return os.WriteFile(s.Path, []byte("[]\n"), 0o600)
}

// Load returns the hosts stored in the JSON file.
func (s *Store) Load() ([]Host, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, fmt.Errorf("read storage file: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return []Host{}, nil
	}

	var hosts []Host
	if err := json.Unmarshal(data, &hosts); err == nil {
		return hosts, nil
	} else {
		var single Host
		if errSingle := json.Unmarshal(data, &single); errSingle != nil {
			return nil, fmt.Errorf("parse storage JSON: %w", err)
		}
		hosts = []Host{single}
		if err := s.Save(hosts); err != nil {
			return nil, fmt.Errorf("migrate storage JSON: %w", err)
		}
		return hosts, nil
	}
}

// Save writes the given hosts to the JSON file.
func (s *Store) Save(hosts []Host) error {
	if err := s.Ensure(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(hosts, "", "  ")
	if err != nil {
		return fmt.Errorf("encode storage JSON: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(s.Path, data, 0o600)
}
