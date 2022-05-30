# Github Actions Runner

![GitHub](https://img.shields.io/github/license/aweris/github-actions-runner)
![GitHub](https://img.shields.io/github/workflow/status/aweris/github-actions-runner/release) [![aweris/github-actions-runner on DockerHub](https://img.shields.io/badge/docker-ready-blue.svg)](https://hub.docker.com/r/aweris/gar) ![GitHub](https://img.shields.io/docker/v/aweris/gar)

Another [self-hosted](https://help.github.com/en/github/automating-your-workflow-with-github-actions/hosting-your-own-runners) Github actions runner.

## Goals

- Auto register and remove
- Support organization level runners
- Support runner self-update

## How to Deploy 

### Docker 

```
docker run aweris/gar:2.292.0 --github-token <pat> --url https://github.com/<your-repo-goes-here>
```

### Helm Chart

Example deployment with helm chart [aweris/gar](https://github.com/aweris/charts). Please check chart repo for configuration options

-  Add Helm repository

```shell
helm repo add aweris https://aweris.github.io/charts/
```

- Update helm repositories

```shell
helm repo update
```

- Create Github Authentication Secret

Create secret using personal access token: 

```shell
kubectl create secret generic github-auth --from-literal=pat=<PAT>
```


Create secret using Github application credentials:
	

```
kubectl create secret generic  github-auth --from-literal=appId=<Github Application ID> \
                                           --from-literal=installationId=<Github Application Installation ID> \
                                           --from-file=privateKey=<Path for the private key file>
```

- Create `values.yaml`

```yaml
runner:
  url: https://github.com/<your-repo-goes-here>
  labels:
    - foo
    - bar
  ghAuth:
    existingSecret: github-auth
```

- Install :

```
helm upgrade --install --values values.yaml runner aweris/gar
```

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

