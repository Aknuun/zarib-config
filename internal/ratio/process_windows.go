//go:build windows

package ratio

import (
	"os"
	"os/exec"
)

func prepareChild(cmd *exec.Cmd, group bool) {}

func signalChild(pid int, group bool, force bool) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}
