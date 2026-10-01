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

// ConnectOptions contains additional options for establishing an SSH connection.
type ConnectOptions struct {
	PortTunnel    string
	ReverseTunnel string
	RemoteCommand string
}

// BuildCommand creates an exec.Cmd for the system ssh client.
// An optional remoteCommand can be passed to execute a command on the remote server
// without opening an interactive session.
func BuildCommand(host storage.Host, remoteCommand ...string) (*exec.Cmd, error) {
	var cmdStr string
	if len(remoteCommand) > 0 {
		cmdStr = remoteCommand[0]
	}
	return BuildCommandWithOptions(host, ConnectOptions{
		RemoteCommand: cmdStr,
	})
}

// BuildCommandWithOptions creates an exec.Cmd for the system ssh client with full options,
// including port tunneling, reverse tunneling, and remote command execution.
func BuildCommandWithOptions(host storage.Host, opts ConnectOptions) (*exec.Cmd, error) {
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

	if opts.PortTunnel != "" {
		formatted, err := FormatPortTunnel(opts.PortTunnel)
		if err != nil {
			return nil, fmt.Errorf("túnel de puerto inválido: %w", err)
		}
		args = append(args, "-L", formatted)
	}

	if opts.ReverseTunnel != "" {
		formatted, err := FormatPortTunnel(opts.ReverseTunnel)
		if err != nil {
			return nil, fmt.Errorf("túnel inverso de puerto inválido: %w", err)
		}
		args = append(args, "-R", formatted)
	}

	args = append(args, host.Address())

	if strings.TrimSpace(opts.RemoteCommand) != "" {
		args = append(args, opts.RemoteCommand)
	}

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

// FormatPortTunnel formats and validates a port tunnel string for ssh -L or ssh -R.
// Supported formats:
//   - "8080"                  -> "8080:localhost:8080"
//   - "8080:80"               -> "8080:localhost:80"
//   - "8080:remotehost:80"    -> "8080:remotehost:80"
//   - "127.0.0.1:8080:h:80"   -> "127.0.0.1:8080:h:80"
func FormatPortTunnel(spec string) (string, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", fmt.Errorf("la especificación de túnel no puede estar vacía")
	}

	parts := strings.Split(spec, ":")
	switch len(parts) {
	case 1:
		if !isValidPort(parts[0]) {
			return "", fmt.Errorf("puerto inválido '%s'", parts[0])
		}
		return fmt.Sprintf("%s:localhost:%s", parts[0], parts[0]), nil
	case 2:
		if !isValidPort(parts[0]) {
			return "", fmt.Errorf("puerto local inválido '%s'", parts[0])
		}
		if !isValidPort(parts[1]) {
			return "", fmt.Errorf("puerto remoto inválido '%s'", parts[1])
		}
		return fmt.Sprintf("%s:localhost:%s", parts[0], parts[1]), nil
	case 3:
		if !isValidPort(parts[0]) {
			return "", fmt.Errorf("puerto local inválido '%s'", parts[0])
		}
		if strings.TrimSpace(parts[1]) == "" {
			return "", fmt.Errorf("host remoto no puede estar vacío")
		}
		if !isValidPort(parts[2]) {
			return "", fmt.Errorf("puerto remoto inválido '%s'", parts[2])
		}
		return spec, nil
	case 4:
		if strings.TrimSpace(parts[0]) == "" {
			return "", fmt.Errorf("ip local de enlace no puede estar vacía")
		}
		if !isValidPort(parts[1]) {
			return "", fmt.Errorf("puerto local inválido '%s'", parts[1])
		}
		if strings.TrimSpace(parts[2]) == "" {
			return "", fmt.Errorf("host remoto no puede estar vacío")
		}
		if !isValidPort(parts[3]) {
			return "", fmt.Errorf("puerto remoto inválido '%s'", parts[3])
		}
		return spec, nil
	default:
		return "", fmt.Errorf("formato no reconocido '%s' (usa formato local:remoto, ej: 8080:80 o 8080)", spec)
	}
}

func isValidPort(p string) bool {
	n, err := strconv.Atoi(p)
	return err == nil && n >= 1 && n <= 65535
}

// BuildSCPCommand creates an exec.Cmd for the system scp client.
func BuildSCPCommand(host storage.Host, rawArgs []string) (*exec.Cmd, error) {
	if len(rawArgs) == 0 {
		return nil, fmt.Errorf("no se especificaron argumentos para scp")
	}

	var scpFlags []string
	var paths []string

	for _, arg := range rawArgs {
		if strings.HasPrefix(arg, "-") && !isRemotePath(arg) {
			scpFlags = append(scpFlags, arg)
		} else {
			paths = append(paths, arg)
		}
	}

	if len(paths) < 2 {
		return nil, fmt.Errorf("scp requiere al menos una ruta de origen y una de destino")
	}

	hasRemote := false
	for _, p := range paths {
		if isRemotePath(p) {
			hasRemote = true
			break
		}
	}

	resolvedPaths := make([]string, len(paths))
	if hasRemote {
		for i, p := range paths {
			if isRemotePath(p) {
				clean := cleanRemotePath(p)
				resolvedPaths[i] = fmt.Sprintf("%s:%s", host.Address(), clean)
			} else {
				resolvedPaths[i] = p
			}
		}
	} else {
		// Convención por defecto: el último argumento es el destino remoto (subida)
		for i := 0; i < len(paths)-1; i++ {
			resolvedPaths[i] = paths[i]
		}
		resolvedPaths[len(paths)-1] = fmt.Sprintf("%s:%s", host.Address(), paths[len(paths)-1])
	}

	port := host.PortOrDefault()
	args := []string{"-P", strconv.Itoa(port)}
	if host.Key != "" {
		resolved, err := ResolveKeyPath(host.Key)
		if err != nil {
			return nil, err
		}
		if resolved != "" {
			args = append(args, "-i", resolved)
		}
	}

	args = append(args, scpFlags...)
	args = append(args, resolvedPaths...)

	cmd := exec.Command("scp", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

// isRemotePath checks whether a path argument represents a remote path on the server.
func isRemotePath(p string) bool {
	if strings.HasPrefix(p, ":") || strings.HasPrefix(p, "remote:") {
		return true
	}
	// Si empieza con letra de unidad en Windows (ej: C:\ o D:/), es local.
	if len(p) >= 3 && p[1] == ':' && (p[2] == '\\' || p[2] == '/') {
		return false
	}
	return false
}

// cleanRemotePath strips remote prefixes like ":" or "remote:".
func cleanRemotePath(p string) string {
	if strings.HasPrefix(p, "remote:") {
		return strings.TrimPrefix(p, "remote:")
	}
	return strings.TrimPrefix(p, ":")
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
