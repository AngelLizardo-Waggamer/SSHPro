package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"ssh-pro/storage"
)

// BuildCommand creates an exec.Cmd for the system ssh client.
func BuildCommand(host storage.Host) (*exec.Cmd, error) {
	port := host.PortOrDefault()
	args := []string{"-p", strconv.Itoa(port)}
	if host.Key != "" {
		resolved, err := ResolveKeyPath(host.Key)
		if err != nil {
			return nil, err
		}
		if resolved != "" {
			args = append(args, "-i", resolved)
		}
	}
	args = append(args, host.Address())

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

// ResolveKeyPath resolves a key path, defaulting to ~/.ssh for relative inputs.
func ResolveKeyPath(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", nil
	}

	key = filepath.FromSlash(key)
	if strings.HasPrefix(key, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		trimmed := strings.TrimPrefix(key, "~")
		trimmed = strings.TrimPrefix(trimmed, string(filepath.Separator))
		trimmed = strings.TrimPrefix(trimmed, "/")
		if trimmed == "" {
			return home, nil
		}
		return filepath.Clean(filepath.Join(home, trimmed)), nil
	}

	if filepath.IsAbs(key) {
		return key, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	sshDir := filepath.Join(home, ".ssh")
	return filepath.Clean(filepath.Join(sshDir, key)), nil
}
