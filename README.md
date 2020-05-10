# Github Actions Runner

Another [self-hosted](https://help.github.com/en/github/automating-your-workflow-with-github-actions/hosting-your-own-runners) Github actions runner.

## Goals

- Auto register and remove
- Support organization level runners
- Support runner self-update

## Usage

### Using executable (with CLI args)

```
Usage of gar:
      --github-token string   Personal access token for authenticate to GitHub
  -l, --labels stringArray    Custom labels for the runner
      --name string           Name of the runner (default "<hostname>")
      --once                  Runner executes only single job
      --replace               Replace any existing runner with the same name (default true)
      --runner-path string    Path of the local runner installation (default "/runner")
      --url string            Repository or Organization url for runner registration
      --version               Prints version info
      --work-dir string       Working directory for the runner (default "/_work")
```

### Environment variables

Environment variables sets property value if it's not set from arguments. CLI arguments has higher priority except `RUNNER_LABELS`.

| Name           | Property         | Description                                      |
|----------------|------------------|--------------------------------------------------|
| GITHUB_TOKEN   | `--github-token` |                                                  |
| REG_URL        | `--url`          |                                                  |
| RUNNER_PATH    | `--runner-path`  |                                                  |
| RUNNER_WORKDIR | `--work-dir`     |                                                  |
| RUNNER_NAME    | `--name`         |                                                  |
| RUNNER_LABELS  | `--labels`       | Comma separated list. Merge values with property |

## Development

```
usage: make [target] ...

targets : 

gar              Builds gar binary
vendor           Updates vendored copy of dependencies
fix              Fix found issues (if it's supported by the $(GOLANGCILINT))
fmt              Runs gofmt
lint             Runs golangci-lint analysis
clean            Cleanup everything
test             Runs go test
install-tools    Install tools
help             Shows this help message
```

## Authors and Acknowledgement

### Inspiration

- [summerwind/actions-runner-controller](https://github.com/summerwind/actions-runner-controller)

