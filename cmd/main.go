package main

import (
	"context"
	"fmt"
	"github.com/aweris/github-actions-runner/runner"
	"github.com/google/go-github/v31/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var (
	ErrMissingParameter = errors.New("missing parameter")
)

func main() {
	var (
		// github client flags
		ghToken string

		// runner flags
		replace      bool
		once         bool
		url          string
		runnerPath   string
		workDir      string
		name         string
		runnerLabels []string
	)

	// github client flags
	pflag.StringVar(&ghToken, "github-token", "", "Personal access token for authenticate to GitHub")

	// runner flags
	pflag.BoolVar(&replace, "replace", true, "Replace any existing runner with the same name")
	pflag.BoolVar(&once, "once", false, "Runner executes only single job")
	pflag.StringVar(&url, "url", "", "Repository or Organization url for runner registration")
	pflag.StringVar(&runnerPath, "runner-path", "/runner", "Path of the local runner installation")
	pflag.StringVar(&workDir, "work-dir", "/_work", "Working directory for the runner")
	pflag.StringVar(&name, "name", getDefaultRunnerName(), "Name of the runner")
	pflag.StringArrayVarP(&runnerLabels, "labels", "l", make([]string, 0), "Custom labels for the runner")

	bindEnv(pflag.Lookup("github-token"), "GITHUB_TOKEN")
	bindEnv(pflag.Lookup("url"), "REG_URL")
	bindEnv(pflag.Lookup("path"), "RUNNER_PATH")
	bindEnv(pflag.Lookup("path"), "RUNNER_WORKDIR")
	bindEnv(pflag.Lookup("name"), "RUNNER_NAME")

	pflag.Parse()

	// merge arg and env labels
	if labels := os.Getenv("RUNNER_LABELS"); len(labels) > 0 {
		runnerLabels = append(runnerLabels, strings.Split(labels, ",")...)
	}

	client, err := NewClientWithAccessToken(ghToken)
	if err != nil {
		log.Fatalf("failed to create GitHub client: %v\n", err)
	}

	tp, err := runner.NewTokenProvider(url, client)
	if err != nil {
		log.Fatalf("failed to create token provider client: %v\n", err)
	}

	// new runner instance
	r, err := runner.NewRunner(
		&runner.Config{
			Replace:       replace,
			Once:          once,
			URL:           url,
			Path:          runnerPath,
			WorkDir:       workDir,
			Name:          name,
			Labels:        runnerLabels,
			TokenProvider: tp,
		},
	)

	if err != nil {
		log.Fatalf("failed to create runner: %v\n", err)
	}

	// graceful shutdown
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c

		log.Printf("system call:%+v", oscall)

		cancel()
	}()

	// start runner
	err = r.Start(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}

// NewClientWithAccessToken returns a client authenticated with personal access token.
func NewClientWithAccessToken(token string) (*github.Client, error) {
	if len(token) == 0 {
		return nil, errors.Wrapf(ErrMissingParameter, "github access token")
	}

	tc := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	return github.NewClient(tc), nil
}

func bindEnv(fn *pflag.Flag, env string) {
	if fn == nil || fn.Changed {
		return
	}

	val := os.Getenv(env)

	if len(val) > 0 {
		if err := fn.Value.Set(val); err != nil {
			log.Fatalf("failed to bind env: %v\n", err)
		}
	}
}

func getDefaultRunnerName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Sprintf("runner-%v", time.Now().UnixNano())
	}

	return hostname
}
