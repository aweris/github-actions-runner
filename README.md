# Github Actions Runner

Another [self-hosted](https://help.github.com/en/github/automating-your-workflow-with-github-actions/hosting-your-own-runners) Github actions runner.

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