package ssh

import (
	"os"
	"os/exec"
	"strconv"

	"ssh-pro/storage"
)

// BuildCommand creates an exec.Cmd for the system ssh client.
func BuildCommand(host storage.Host) *exec.Cmd {
	port := host.PortOrDefault()
	args := []string{"-p", strconv.Itoa(port)}
	if host.Key != "" {
		args = append(args, "-i", host.Key)
	}
	args = append(args, host.Address())

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
