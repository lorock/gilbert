package build

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-gilbert/gilbert-sdk"

	"github.com/go-gilbert/gilbert/support/shell"
)

// Action represents Gilbert's plugin
type Action struct {
	scope  sdk.ScopeAccessor
	cmd    *exec.Cmd
	params Params
}

// Call calls a plugin
func (a *Action) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) (err error) {
	a.cmd, err = a.params.newCompilerProcess(a.scope)
	if err != nil {
		return err
	}

	ctx.Log().Debugf("Target: %s %s", a.params.Target.Os, a.params.Target.Arch)
	ctx.Log().Debugf("Command: '%s'", strings.Join(a.cmd.Args, " "))
	a.cmd.Stdout = ctx.Log()
	a.cmd.Stderr = ctx.Log().ErrorWriter()

	if err := a.cmd.Start(); err != nil {
		return fmt.Errorf(`failed build project, %s`, err)
	}

	if err = a.cmd.Wait(); err != nil {
		return shell.FormatExitError(err)
	}

	return nil
}

// Cancel cancels build process
func (a *Action) Cancel(ctx sdk.JobContextAccessor) error {
	if a.cmd != nil {
		if err := shell.KillProcessGroup(a.cmd); err != nil {
			ctx.Log().Debug(err.Error())
		}
	}

	return nil
}
