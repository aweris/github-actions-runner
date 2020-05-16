package main

import (
	"context"
	"fmt"
	"github.com/aweris/github-actions-runner/runner"
	"github.com/google/go-github/v31/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/spf13/pflag"
)

var (
	// version flags
	version = "dev"
	commit  = "none"
	date    = "unknown"

	// Errors
	ErrMissingParameter = errors.New("missing parameter")
)

func main() {
	var (
		// github client flags
		ghToken             string // github personal access token
		ghAppID             int64  // github application id
		ghAppInstallationID int64  // github application installation id
		ghAppPrivateKeyPath string // github application key path

		// runner flags
		replace      bool
		once         bool
		url          string
		runnerPath   string
		workDir      string
		name         string
		runnerLabels []string

		// other
		showVersion bool
	)

	// github client flags
	pflag.StringVar(&ghToken, "github-token", "", "Personal access token for authenticate to GitHub")
	pflag.Int64Var(&ghAppID, "github-app-id", 0, "Github application ID")
	pflag.Int64Var(&ghAppInstallationID, "github-app-installation-id", 0, "Installation ID for the Github application")
	pflag.StringVar(&ghAppPrivateKeyPath, "github-app-private-key", "", "The path of a private key file to authenticate as a GitHub App")

	// runner flags
	pflag.BoolVar(&showVersion, "version", false, "Prints version info")
	pflag.BoolVar(&replace, "replace", true, "Replace any existing runner with the same name")
	pflag.BoolVar(&once, "once", false, "Runner executes only single job")
	pflag.StringVar(&url, "url", "", "Repository or Organization url for runner registration")
	pflag.StringVar(&runnerPath, "runner-path", "/runner", "Path of the local runner installation")
	pflag.StringVar(&workDir, "work-dir", "/_work", "Working directory for the runner")
	pflag.StringVar(&name, "name", getDefaultRunnerName(), "Name of the runner")
	pflag.StringArrayVarP(&runnerLabels, "labels", "l", make([]string, 0), "Custom labels for the runner")

	bindEnv(pflag.Lookup("github-token"), "GITHUB_TOKEN")
	bindEnv(pflag.Lookup("github-app-id"), "GITHUB_APP_ID")
	bindEnv(pflag.Lookup("github-app-installation-id"), "GITHUB_APP_INSTALLATION_ID")
	bindEnv(pflag.Lookup("github-app-private-key"), "GITHUB_APP_PRIVATE_KEY_PATH")
	bindEnv(pflag.Lookup("url"), "REG_URL")
	bindEnv(pflag.Lookup("path"), "RUNNER_PATH")
	bindEnv(pflag.Lookup("path"), "RUNNER_WORKDIR")
	bindEnv(pflag.Lookup("name"), "RUNNER_NAME")

	pflag.Parse()

	if showVersion {
		fmt.Printf("Version    : %s\n", version)
		fmt.Printf("Git Commit : %s\n", commit)
		fmt.Printf("Build Date : %s\n", date)
		os.Exit(0)
	}

	// merge arg and env labels
	if labels := os.Getenv("RUNNER_LABELS"); len(labels) > 0 {
		runnerLabels = append(runnerLabels, strings.Split(labels, ",")...)
	}

	if ghToken == "" && ghAppID == 0 {
		log.Fatal("please provide personal access token or github application credentials")
	}

	var client *github.Client
	var err error

	if ghToken != "" {
		client, err = NewClientWithAccessToken(ghToken)
		if err != nil {
			log.Fatalf("failed to create GitHub client: %v\n", err)
		}
	} else {
		if ghAppInstallationID == 0 {
			log.Fatal("missing application installation id")
		}

		if ghAppPrivateKeyPath == "" {
			log.Fatal("missing application private key path")
		}

		client, err = NewClient(ghAppID, ghAppInstallationID, ghAppPrivateKeyPath)
		if err != nil {
			log.Fatalf("failed to create GitHub client: %v\n", err)
		}
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

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

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

// NewClient returns a client authenticated as a GitHub App.
func NewClient(appID, installationID int64, privateKeyPath string) (*github.Client, error) {
	tr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	return github.NewClient(&http.Client{Transport: tr}), nil
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
