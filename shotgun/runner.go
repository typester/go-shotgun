package shotgun

import (
	"errors"
	"os"
	"os/exec"
	"sync"
	"time"

	"fmt"
	"syscall"
)

type Runner struct {
	mu          sync.Mutex
	cmd         *exec.Cmd
	command     string
	args        []string
	needRestart bool
}

func NewRunner(cmd []string) (*Runner, error) {
	if len(cmd) < 1 {
		return nil, errors.New("command required")
	}

	return &Runner{command: cmd[0], args: cmd[1:]}, nil
}

func (r *Runner) Start() error {
	r.cmd = exec.Command(r.command, r.args...)
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	return r.cmd.Start()
}

func (r *Runner) Kill() error {
	return r.cmd.Process.Kill()
}

func (r *Runner) Terminate() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return errors.New("Couldn't terminate process that is not running")
	}

	fmt.Printf("shutdown app...\n")

	timeout := time.After(10 * time.Second)
	quit := make(chan bool)

	go func() {
		r.cmd.Process.Wait()
		quit <- true
	}()

	err := r.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}

	select {
	case <-timeout:
		fmt.Fprintf(os.Stderr, "timeout waiting process end, nowforce Kill it\n")
		err = r.Kill()
	case <-quit:
	}

	return err
}

func (r *Runner) SetNeedRestart() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.needRestart = true
}

func (r *Runner) CheckRestart() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// program not started
	if r.cmd == nil {
		return r.Start()
	}

	if r.needRestart {
		r.needRestart = false
		r.Terminate()
		return r.Start()
	}

	return nil
}
