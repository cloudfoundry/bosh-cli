## Unit Tests

Each package in the agent has its own unit tests and there are integration tests in the `integration` package.

You can also run all tests with `bin/test`.


## Acceptance Tests

The acceptance tests are designed to exercise the main commands of the bosh-micro CLI (deployment, deploy, delete). 

They are not designed to verify the compatibility of CPIs or testing BOSH releases. 

### Dependencies

- [Vagrant](https://www.vagrantup.com/)

    `brew install vagrant`

- [sshpass](http://linux.die.net/man/1/sshpass)

    `brew install https://raw.github.com/eugeneoden/homebrew/eca9de1/Library/Formula/sshpass.rb`

### Providers

Acceptance tests can be run in a VM with the following vagrant providers:

* [virtualbox](https://www.virtualbox.org/) (free)
* [vmware_fusion](http://www.vmware.com/products/fusion)
* [vmware_workstation](http://www.vmware.com/products/workstation)
* [aws](http://aws.amazon.com/)

#### Local Provider

The acceptance tests can be run on a local VM (using virtual box, vmware fusion, or vmware_workstation with vagrant).

The acceptance tests require a stemcell and a BOSH Warden CPI release.
  
Without specifying them, a specific (known to work) version of each will be downloaded.
  
You may alternatively choose to download them to a local directory and specify their paths via environment variables.
They will then be scp'd onto the vagrant VM.
  
To take advantage of this feature, export the following variables prior to running the tests (absolute paths are required):

```
$ export BOSH_MICRO_CPI_RELEASE_PATH=/tmp/bosh-warden-cpi-9.tgz
$ export BOSH_MICRO_STEMCELL_PATH=/tmp/bosh-stemcell-348-warden-boshlite-ubuntu-trusty-go_agent.tgz
$ ./bin/test-acceptance-with-vm --provider=virtualbox
```

#### AWS Provider

The acceptance tests can also be run on a remote VM (using aws with vagrant).

When using the AWS provider, you will need to provide the following:

```
$ export BOSH_MICRO_PRIVATE_KEY=/tmp/id_rsa

# The following variables are required by Bosh Lite
$ export BOSH_AWS_ACCESS_KEY_ID=access_key
$ export BOSH_AWS_SECRET_ACCESS_KEY=secret
$ export BOSH_LITE_KEYPAIR=keypair
$ export BOSH_LITE_SECURITY_GROUP=sg-1234
$ export BOSH_LITE_PRIVATE_KEY=$BOSH_MICRO_PRIVATE_KEY
```

#### Running tests against existing VM

Acceptance tests use configuration file specified via `BOSH_MICRO_CONFIG_PATH`. The format of the configuration file is basic JSON.

```
{
  "vm_username": "TEST_VM_USERNAME",
  "vm_ip": "TEST_VM_IP",
  "private_key_path": "TEST_VM_PRIVATE_KEY_PATH",
  "cpi_release_path": "CPI_RELEASE_PATH",
  "cpi_release_url": "CPI_RELEASE_URL",
  "stemcell_path": "STEMCELL_PATH",
  "stemcell_url": "STEMCELL_URL"
  "dummy_release_path": "DUMMY_RELEASE_PATH",
}
```

Run acceptance tests:

```
BOSH_MICRO_CONFIG_PATH=config.json bin/test-acceptance
```

## Debugging Acceptance Test Failures

If your acceptance tests are failing mysteriously while running a command, here are some things to check:

 * `vagrant ssh` to the vm running the specs and check out the `bosh-micro-cli.log` in the vagrant user home directory
 * If your agent isn't starting, get its IP from the micro logs (see above). Then you can `ssh vcap@<ip>` and check out `/var/vcap/bosh/log/current`
