# BOSH CLI

* Documentation: [bosh.io/docs](https://bosh.io/docs)
* Slack: #bosh on <https://slack.cloudfoundry.org>
* Mailing list: [cf-bosh](https://lists.cloudfoundry.org/pipermail/cf-bosh)
* CI: <https://main.bosh-ci.cf-app.com/pipelines/bosh:cli>
* Roadmap: [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/956238)

## Usage

Relevant documentation pages from bosh.io:

- [Installing BOSH](https://bosh.io/docs#install)

## Client Library

This project includes [`director`](director/interfaces.go) and [`uaa`](uaa/interfaces.go) packages meant to be used in your project for programmatic access to the Director API.

See [docs/example.go](docs/example.go) for a live short usage example.

## YAML Encoding Issues

Some older versions of the ruby `bosh` CLI wrote non-standard YAML files which this CLI cannot parse. You can migrate those files with the following command and commit them to your repository. They will continue to work with older ruby `bosh` CLI.

    $ cd your-release-directory
    $ ruby -e 'require "YAML"; `grep -rl "binary |-" releases/*.yml .final_builds`.split("\n").each{|file| yaml = YAML.load(File.read(file)); File.write(file, YAML.dump(yaml)) }'

## Developer Notes

- [Workstation setup docs](docs/build.md)
- [Test docs](docs/test.md)
- [CLI workflow](docs/cli_workflow.md)
  - [Architecture docs](docs/architecture.md)
