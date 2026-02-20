package platform

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/proc"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/subprocess/command"
)

// RemoveWithPrivileges removes the file or directory at path using elevated
// privileges (equivalent to `sudo rm -rf <path>`). The call delegates removal
// to helper functions that execute the removal as a privileged user.
//
// Behaviour
//   - If path is empty, the function returns nil and does nothing.
//   - The function performs a minimal sanitisation of path with filepath.Clean()
//     to normalise separators and collapse `.`/`..` elements. This is intended
//     to reduce accidental path traversal surprises but does not enforce any
//     allow-listing or other policy.
//
// Security notes
//   - This function requires that the current process (or the actor used by
//     WithPrivileges) has appropriate privileges to remove the target path.
//   - It intentionally performs no access-control checks, host-level policy or
//     allow-listing. Callers are responsible for ensuring that the supplied
//     path is expected and authorised for privileged removal.
//   - Static analysis tools (gosec) flag uses of os.Stat/os.Remove on
//     externally-supplied paths as potential path-traversal vulnerabilities
//     (G703). The function deliberately delegates responsibility for such
//     validation to callers and therefore uses a targeted suppression with a
//     clear justification below. If you prefer centralised validation,
//     implement caller-side checks (for example: enforce base directory,
//     enforce absolute paths, or test against an allow-list) prior to calling
//     this function.
func RemoveWithPrivileges(ctx context.Context, path string) (err error) {
	if reflection.IsEmpty(path) {
		return
	}

	// Normalise the path to collapse any . or .. segments and remove
	// redundant separators. This reduces the chance of accidental
	// path-traversal surprises while preserving intended semantics.
	path = filepath.Clean(path)

	// The following os.Stat call inspects the (cleaned) path before removal.
	// gosec G703 may warn about path traversal here; this is intentional:
	// the function purposefully performs privileged removals and delegates
	// policy/authorisation to the caller (see docstring). Suppress G703 with
	// a clear justification rather than hiding the finding.
	fi, err := os.Stat(path) //nolint:gosec //G703 // Reason: deliberate privileged removal; caller must validate path/authorisation
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrNotFound, err, "could not find path [%v]", path)
		return
	}

	if fi.IsDir() {
		err = removeDirAs(ctx, WithPrivileges(command.Me()), path)
	} else {
		err = removeFileAs(ctx, WithPrivileges(command.Me()), path)
	}
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not remove the path [%v]", path)
	}
	return
}

func executeCommandAs(ctx context.Context, as *command.CommandAsDifferentUser, args ...string) error {
	if as == nil {
		return commonerrors.UndefinedVariable("command wrapper")
	}
	if len(args) == 0 {
		return commonerrors.UndefinedVariable("command to execute")
	}
	cmdName, cmdArgs := as.RedefineCommand(args...)
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	// setting the following to avoid having hanging subprocesses as described in https://github.com/golang/go/issues/24050
	cmd.WaitDelay = 5 * time.Second
	cmd, err := proc.DefineCmdCancel(cmd)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not set the command cancel function")
	}
	return runCommand(args[0], cmd)
}

func executeCommand(ctx context.Context, args ...string) error {
	return executeCommandAs(ctx, WithPrivileges(command.Me()), args...)
}

func runCommand(cmdDescription string, cmd *exec.Cmd) error {
	_, err := cmd.Output()
	if err == nil {
		return nil
	}
	newErr := commonerrors.ConvertContextError(err)
	switch {
	case commonerrors.Any(newErr, commonerrors.ErrTimeout, commonerrors.ErrCancelled, commonerrors.ErrUnknown, commonerrors.ErrUnsupported, commonerrors.ErrCondition, commonerrors.ErrForbidden):
		return newErr
	default:
		details := "no further details"
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			details = string(exitError.Stderr)
		}
		newErr = commonerrors.WrapErrorf(commonerrors.ErrUnknown, err, "the command `%v` failed (%v)", cmdDescription, details)
		return newErr
	}
}
