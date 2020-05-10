package runner

import (
	"context"
	"log"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// Config represents runner configuration
type Config struct {
	Replace       bool           // config.sh --replace option. default true
	Once          bool           // run.sh --once option
	URL           string         // organization/repository registration url
	Path          string         // installation path of the runner.
	WorkDir       string         // working directory for the runner
	Name          string         // name of the runner
	Labels        []string       // additional runner labels
	TokenProvider *TokenProvider // token provider instance
}

// Runner represents self-hosted runner instance
type Runner struct {
	registered bool     // is runner already registered the Github
	config     *Config  // runner config
	runCMD     *Command // run.sh command
	configCMD  *Command // config.sh command
}

// NewRunner returns new Runner instance
func NewRunner(config *Config) (*Runner, error) {
	// Configure runner
	runCmd, err := NewCommand(path.Join(config.Path, "run.sh"), os.Stdout, os.Stderr)
	if err != nil {
		return nil, err
	}

	configCMD, err := NewCommand(path.Join(config.Path, "config.sh"), os.Stdout, os.Stderr)
	if err != nil {
		return nil, err
	}

	// it means runner already registered
	registered := isPathExist(path.Join(config.Path, ".credentials")) && isPathExist(path.Join(config.Path, ".runner"))

	return &Runner{
		registered: registered,
		config:     config,
		runCMD:     runCmd,
		configCMD:  configCMD,
	}, nil
}

func (r *Runner) Start(ctx context.Context) error {
	// register runner
	err := r.register()
	if err != nil {
		return err
	}

	// remove runner before finish process
	defer func() {
		err := r.remove()
		if err != nil {
			log.Fatalf("failed to remove runner from github err: %v", err)
		}
	}()

	args := make([]string, 0)

	if r.config.Once {
		args = append(args, "--once")
	}

	err = r.runCMD.Run(ctx, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runner) register() error {
	// it means already registered and don't want to override current config
	if r.registered {
		return nil
	}

	config := r.config

	args := make([]string, 0)

	args = append(args, "--unattended", "--url", config.URL, "--name", config.Name)

	if len(config.WorkDir) > 0 {
		args = append(args, "--work", config.WorkDir)
	}

	if config.Replace {
		args = append(args, "--replace")
	}

	if len(config.Labels) > 0 {
		args = append(args, "--labels", strings.Join(config.Labels, ","))
	}

	tok, err := config.TokenProvider.CreateRegistrationToken()
	if err != nil {
		return err
	}

	args = append(args, "--token", *tok.Token)

	// configure the runner
	err = r.configCMD.Run(context.Background(), args...)
	if err != nil {
		return err
	}

	// mark registration flag true
	r.registered = true

	return nil
}

func (r *Runner) remove() error {
	// No need to remove since it's not registered
	if !r.registered {
		return nil
	}

	args := make([]string, 0)

	args = append(args, "remove")

	tok, err := r.config.TokenProvider.CreateRemoveToken()
	if err != nil {
		return err
	}

	args = append(args, "--token", *tok.Token)

	err = r.configCMD.Run(context.Background(), args...)
	if err != nil {
		return errors.Wrap(err, "failed to remove runner")
	}

	// Update registration flag
	r.registered = false

	return nil
}

func isPathExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
