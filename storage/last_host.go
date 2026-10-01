package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LastHostPath returns the path to the file storing the last connected host name.
func LastHostPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".sshpro", "last_host"), nil
}

// SaveLastHost writes the name of the last connected host to disk.
func SaveLastHost(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	path, err := LastHostPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create storage dir: %w", err)
	}

	return os.WriteFile(path, []byte(name+"\n"), 0o600)
}

// GetLastHost reads the name of the last connected host from disk.
func GetLastHost() (string, error) {
	path, err := LastHostPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read last host file: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}
