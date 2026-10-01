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

// FindPublicKey locates and reads a public SSH key.
// If explicitPath is provided, it attempts to load that key.
// Otherwise, it checks for a .pub matching host.Key, or common defaults in ~/.ssh/.
func FindPublicKey(explicitPath string, host storage.Host) (string, string, error) {
	if strings.TrimSpace(explicitPath) != "" {
		resolved, err := ResolveKeyPath(explicitPath)
		if err != nil {
			return "", "", fmt.Errorf("error al resolver la ruta de la clave pública: %w", err)
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			return "", "", fmt.Errorf("no se pudo leer la clave pública en '%s': %w", resolved, err)
		}
		return resolved, strings.TrimSpace(string(data)), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("no se pudo resolver el directorio home del usuario: %w", err)
	}
	sshDir := filepath.Join(home, ".ssh")

	var candidates []string

	if strings.TrimSpace(host.Key) != "" {
		key := strings.TrimSpace(host.Key)
		if strings.HasSuffix(key, ".pub") {
			candidates = append(candidates, key)
		} else {
			candidates = append(candidates, key+".pub")
		}
	}

	candidates = append(candidates,
		"id_ed25519.pub",
		"id_ecdsa.pub",
		"id_rsa.pub",
		"id_dsa.pub",
	)

	for _, cand := range candidates {
		var fullPath string
		if filepath.IsAbs(cand) || strings.HasPrefix(cand, "~") {
			resolved, err := ResolveKeyPath(cand)
			if err != nil {
				continue
			}
			fullPath = resolved
		} else {
			fullPath = filepath.Join(sshDir, cand)
		}

		if data, err := os.ReadFile(fullPath); err == nil && len(data) > 0 {
			return fullPath, strings.TrimSpace(string(data)), nil
		}
	}

	return "", "", fmt.Errorf("no se encontró ninguna clave pública (.pub) en ~/.ssh/ ni asociada al host")
}

// BuildCopyIDCommand creates an exec.Cmd that appends the given public key to
// ~/.ssh/authorized_keys on the remote server using the ssh client.
func BuildCopyIDCommand(host storage.Host, pubKey string) (*exec.Cmd, error) {
	port := host.PortOrDefault()
	args := []string{"-p", strconv.Itoa(port)}

	args = append(args, host.Address())

	remoteCommand := "umask 077; test -d ~/.ssh || mkdir -p ~/.ssh; cat >> ~/.ssh/authorized_keys"
	args = append(args, remoteCommand)

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = strings.NewReader(strings.TrimSpace(pubKey) + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}
