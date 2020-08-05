package runner

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

const (
	AutoUpdateExitCode int = 3
)

// Command is a simple wrapper for exec.CommandContext
type Command struct {
	path   string
	stdout io.Writer
	stderr io.Writer
}

func NewCommand(path string, stdout, stderr io.Writer) (*Command, error) {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "invalid command path %s", path)
	}

	if !isExecAny(info.Mode()) {
		return nil, errors.Wrapf(err, "command is not executable %s", path)
	}

	return &Command{
		path:   path,
		stdout: stdout,
		stderr: stderr,
	}, nil
}

func (c *Command) Run(ctx context.Context, args ...string) error {
	// ignore lint for: `G204: Subprocess launched with function call as argument or cmd arguments (gosec)`
	cmd := exec.CommandContext(ctx, c.path, args...) //nolint:gosec

	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 3 means auto update for runner
			if exitError.ExitCode() == AutoUpdateExitCode {
				logrus.Warningln("!!!! Runner Auto Upgrade !!!!")
				return c.Run(ctx, args...)
			}
		}

		return err
	}

	return nil
}

func isExecAny(mode os.FileMode) bool {
	return mode&0111 != 0
}
