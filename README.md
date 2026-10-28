# bosh CLI

* Documentation: [bosh.io/docs/cli-v2](https://bosh.io/docs/cli-v2.html)
* Slack: #bosh on <https://slack.cloudfoundry.org>
* Mailing list: [cf-bosh](https://lists.cloudfoundry.org/pipermail/cf-bosh)
* CI: <https://main.bosh-ci.cf-app.com/teams/main/pipelines/bosh:cli>
* Roadmap: [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/956238)

## Install

- [https://bosh.io/docs/cli-v2-install/](https://bosh.io/docs/cli-v2-install/)

## Client Library

This project includes [`director`](director/interfaces.go) and [`uaa`](uaa/interfaces.go) packages meant to be used in your project for programmatic access to the Director API.

See [docs/example.go](docs/example.go) for a live short usage example.

## Developer Notes

- [Workstation setup docs](docs/build.md)
- [Test docs](docs/test.md)
- [CLI workflow](docs/cli_workflow.md)
  - [Architecture docs](docs/architecture.md)

## Release archive limits

Release, job, package and license extraction share a process-wide budget. The
byte limit defaults to the smaller of 128 GiB and 80% of available temporary-disk
space; the member limit defaults to 10,000 and shrinks on nearly full filesystems.
Manifest contents share a 32 MiB budget so subsequent YAML parsing cannot read an
unbounded member, including through archive links. Limits apply while streaming,
before a member is written. Tar metadata and
trailing compressed data are also bounded. XZ and Zstandard archives require the
corresponding system decoder and allow at most 256 MiB of decoder memory.

For a larger trusted release, set `BOSH_RELEASE_MAX_EXPANDED_BYTES` to a positive
byte count and `BOSH_RELEASE_MAX_ARCHIVE_MEMBERS` to a positive count before
running the CLI. For example, 274877906944 permits 256 GiB of file contents.
Overrides replace the defaults; ensure the temporary filesystem has enough free
space for the selected limits.

`BOSH_RELEASE_MAX_MANIFEST_BYTES` overrides the combined manifest byte limit.
