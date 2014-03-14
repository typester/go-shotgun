// +build !windows

package shotgun

import (
	"os"
	"os/exec"
	"syscall"
)

func (r *Runner) Start() error {
	r.cmd = exec.Command(r.command, r.args...)
	r.cmd.Stdin = nil
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	return r.cmd.Start()
}

func (r *Runner) Signal() error {
	err := r.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	return nil
}
