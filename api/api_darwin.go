package api

import (
	"os/exec"
	"syscall"
)

func setProcessGroup(cmd *exec.Cmd) {
	// macOS doesn't support Pdeathsig, but we can set Setpgid
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
