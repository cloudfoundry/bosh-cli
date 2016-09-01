# bosh CLI

* Documentation: [bosh.io/docs](https://bosh.io/docs)
* Slack: #bosh on <https://slack.cloudfoundry.org>
* Mailing list: [cf-bosh](https://lists.cloudfoundry.org/pipermail/cf-bosh)
* CI: <https://main.bosh-ci.cf-app.com/pipelines/bosh:cli>
* Roadmap: [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/956238)

## Usage

Relevant documentation pages from bosh.io:

- [Installing BOSH](https://bosh.io/docs#install)

To install the very latest `bosh2` CLI on OS X (assuming `~/bin` is in `$PATH`): 

```
version=$(curl -s https://api.github.com/repos/cloudfoundry/bosh-cli/releases/latest | jq -r .tag_name | sed -e "s/^v//") && wget https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-$version-darwin-amd64 && mv bosh-cli-$version-darwin-amd64 ~/bin/bosh2 && chmod +x ~/bin/bosh2
```


## Client Library

This project includes [`director`](director/interfaces.go) and [`uaa`](uaa/interfaces.go) packages meant to be used in your project for programmatic access to the Director API.

See [docs/example.go](docs/example.go) for a live short usage example.

## Developer Notes

- [Workstation setup docs](docs/build.md)
- [Test docs](docs/test.md)
- [CLI workflow](docs/cli_workflow.md)
  - [Architecture docs](docs/architecture.md)
