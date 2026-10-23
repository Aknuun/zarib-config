//go:build !windows

package ratio

import (
	"os/exec"
	"syscall"
)

func prepareChild(cmd *exec.Cmd, group bool) {
	if group {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
}

func signalChild(pid int, group bool, force bool) error {
	signal := syscall.SIGTERM
	if force {
		signal = syscall.SIGKILL
	}
	if group {
		return syscall.Kill(-pid, signal)
	}
	return syscall.Kill(pid, signal)
}
