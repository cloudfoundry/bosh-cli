# bosh-init

`bosh-init` is a tool used to create and update the Director (its VM and persistent disk) in a BOSH environment.

* Documentation: [bosh.io/docs](https://bosh.io/docs)
* IRC: [`#bosh` on freenode](http://webchat.freenode.net/?channels=bosh)
* Mailing list: [cf-bosh](https://lists.cloudfoundry.org/pipermail/cf-bosh)
* CI: [https://main.bosh-ci.cf-app.com/pipelines/bosh-init](https://main.bosh-ci.cf-app.com/pipelines/bosh-init)
* Roadmap: [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/1133984)

## Usage

Relevant documentation pages from bosh.io:

- [Installing BOSH](http://bosh.io/docs#install)
- [Install bosh-init](https://bosh.io/docs/install-bosh-init.html)
- [Using bosh-init](https://bosh.io/docs/using-bosh-init.html)

## Developer Notes

See the [CLI workflow](docs/cli_workflow.md) for more information on creating a manifest.

To build bosh-init see our [workstation setup docs](https://github.com/cloudfoundry/bosh-init/blob/master/docs/build.md).

To run bosh-init tests see our [test docs](https://github.com/cloudfoundry/bosh-init/blob/master/docs/test.md).

To deploy BOSH with UAA using bosh-init see our [UAA docs](https://github.com/cloudfoundry/bosh-init/blob/master/docs/uaa.md).

To learn more about the bosh-init design see our [architecture docs](https://github.com/cloudfoundry/bosh-init/blob/master/docs/architecture.md).
