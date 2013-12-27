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
	r.cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_UNICODE_ENVIRONMENT | 0x00000200,
	}
	return r.cmd.Start()
}

func (r *Runner) Signal() error {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	defer dll.Release()
	f, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	pid := r.cmd.Process.Pid
	r1, _, err := f.Call(uintptr(syscall.CTRL_BREAK_EVENT), uintptr(pid))
	if r1 == 0 {
		return err
	}
	return nil
}
