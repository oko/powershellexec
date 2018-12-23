package powershellexec

import (
	"testing"
)

func TestExecBasic(t *testing.T) {
	var exe Executor
	exe = &WrappedExecutor{}
	_, _, err := exe.Execute("whoami")
	if err != nil {
		t.Errorf("PS `whoami` failed")
	}
}

func TestWrappedExecutor(t *testing.T) {
	var exe Executor
	var err error
	exe = &WrappedExecutor{}
	exe.SetExitCodes([]int{1, 2, 3})
	_, _, err = exe.Execute("exit 0")
	if err != nil {
		t.Errorf("PS `exit 0` errored")
	}
	_, _, err = exe.Execute("exit 1")
	if err != nil {
		t.Errorf("PS `exit 1` errored")
	}
	_, _, err = exe.Execute("exit 2")
	if err != nil {
		t.Errorf("PS `exit 2` errored")
	}
	_, _, err = exe.Execute("exit 3")
	if err != nil {
		t.Errorf("PS `exit 3` errored")
	}
	_, _, err = exe.Execute("exit 4")
	if err == nil {
		t.Errorf("PS `exit 4` did not error but should have")
	}
}
