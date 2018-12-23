package powershellexec

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// ExecutionError is an error type for errors that happen in os/exec
// commands between cmd.Start() and cmd.Wait() return
type ExecutionError struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	InnerErr error
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf(
		"execution error: %s %s exit code %d inner err %s",
		string(e.Stdout), string(e.Stderr), e.ExitCode, e.InnerErr,
	)
}

// Executor defines the execution interface for PS runners
type Executor interface {
	SetExitCodes([]int)
	Execute(string) ([]byte, []byte, error)
}

// WrappedExecutor defines a powershell executor that runs its scripts
// inside a wrapper script (from github.com/chef/chef)
type WrappedExecutor struct {
	okExitCodes []int
}

func (w *WrappedExecutor) writeTempScript(dst *os.File, script string) error {
	gen := strings.Replace(PowershellWrapperScript, "SCRIPTBLOCK", script, 1)
	_, err := dst.Write([]byte(gen))
	if err != nil {
		return err
	}
	return nil
}

// SetExitCodes sets acceptable exit codes for scripts run by
// this executor (besides exit 0)
func (w *WrappedExecutor) SetExitCodes(newCodes []int) {
	w.okExitCodes = newCodes
}

// Execute actually executes Powershell scripts, as well as capturing
// output and exit errors.
func (w *WrappedExecutor) Execute(script string) (stdout, stderr []byte, err error) {
	localExits := make([]int, len(w.okExitCodes))
	copy(localExits, w.okExitCodes)

	tmpf, err := ioutil.TempFile("", "pswrapper.*.ps1")
	if err != nil {
		return nil, nil, err
	}
	w.writeTempScript(tmpf, script)
	tmpf.Close()

	args := []string{
		"-NoLogo",
		"-NonInteractive",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-InputFormat", "None",
		"-File", tmpf.Name(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell.exe", args...)

	var bStdout, bStderr bytes.Buffer
	cmd.Stdout = &bStdout
	cmd.Stderr = &bStderr
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	err = cmd.Wait()
	log.Printf("error: %s", err)
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				sc := status.ExitStatus()
				for _, ec := range localExits {
					if sc == ec {
						err = nil
					}
				}
				if err != nil {
					err = &ExecutionError{
						Stdout:   bStdout.Bytes(),
						Stderr:   bStderr.Bytes(),
						ExitCode: sc,
						InnerErr: nil,
					}
				}
			}
		}
	}
	return bStdout.Bytes(), bStderr.Bytes(), err
}
