# `powershellexec`
This package provides an interface for executing Powershell commands from Go code.

## Usage
Run a simple Powershell command:

```
package main

import (
	"github.com/oko/powershellexec"
	"log"
)

func main() {
	exe := &powershellexec.WrappedExecutor{}
	_, _, err := exe.Execute("whoami")
	if err != nil {
		log.Fatalf("failed to run `whoami` in Powershell: %s", err)
	}
}
```

If your command will return unusual exit codes (i.e., `robocopy` calls):

```
exe.SetExitCodes([]int{1,2,3})
```

The array of exit codes passed will be checked at command completion.
