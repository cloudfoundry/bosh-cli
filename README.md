# BOSH Micro CLI [![Build Status](https://travis-ci.org/cloudfoundry/bosh-micro-cli.svg?branch=master)](https://travis-ci.org/cloudfoundry/bosh-micro-cli)

This is the BOSH Micro CLI rewritten to support external CPIs.

* Documentation: [bosh.io/docs](https://bosh.io/docs)
* IRC: `#bosh` on freenode
* Google groups:
  [bosh-users](https://groups.google.com/a/cloudfoundry.org/group/bosh-users/topics) &
  [bosh-dev](https://groups.google.com/a/cloudfoundry.org/group/bosh-dev/topics) &
  [vcap-dev](https://groups.google.com/a/cloudfoundry.org/group/vcap-dev/topics) (for CF)
* Roadmap: [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/1133984)

## Usage

1. Build micro:

  ```
  bin/build
  ```

2. Set up deployment manifest:

  ```
  out/bosh-micro deployment manifest.yml
  ```

3. Deploy

  ```
  out/bosh-micro deploy stemcell.tgz cpi-release.tgz
  ```

where `cpi-release.tgz` is a BOSH CPI release and `stemcell.tgz` is a BOSH stemcell appropriate for the CPI release.

Please see the [CLI workflow](docs/cli_workflow.md) for more information on creating a manifest.

## Logging

To output logs during bosh-micro commands set the `BOSH_MICRO_LOG_LEVEL` environment variable to one of the following values: 

DEBUG, INFO, WARN, ERROR, NONE (default)

To output logs to a file set the `BOSH_MICRO_LOG_PATH` environment variable to the path of the file to create and/or append to. 

By default (when `BOSH_MICRO_LOG_LEVEL` is not NONE) logs write to STDOUT (debug & info) & STDERR (warn & error).


## Contributing

### Set up a workstation for development

Note: This guide assumes a few things:

- You have gcc (or an equivalent)
- You can install packages (brew, apt-get, or equivalent)

Get Golang and its dependencies (Mac example, replace with your package manager of choice):

- `brew update`
- `brew install go`
- `brew install git` (Go needs git for the `go get` command)
- `brew install hg` (Go needs mercurial for the `go get` command)

Clone and set up the BOSH Micro CLI repository:

- `go get -d github.com/cloudfoundry/bosh-micro-cli`
- `cd $GOPATH/src/github.com/cloudfoundry/bosh-micro-cli`

From here on out we assume you're working in `$GOPATH/src/github.com/cloudfoundry/bosh-micro-cli`

To build the micro cli:

- `bin/build` # The `bosh-micro` binary will be located in `out/`

Install tools used by the BOSH Micro CLI test suite:

- `bin/go get code.google.com/p/go.tools/cmd/vet`
- `bin/go get github.com/golang/lint/golint`

### Running unit tests

Each package in the agent has its own unit tests and there are integration tests in the `integration` package.
You can also run all tests with `bin/test`.

### Running acceptance tests

Install sshpass:

- `brew install https://raw.github.com/eugeneoden/homebrew/eca9de1/Library/Formula/sshpass.rb`

Acceptance tests can be run in a VM with the following vagrant providers:

* virtualbox
* vmware_fusion
* vmware_workstation
* aws

#### Running tests using local provider

  When using a local provider, you may choose to download both the
  bosh-warden-cpi-release and stemcell to a local directory to then be copied
  into the VM. To take advantage of this feature, export the following variables
  prior to running the tests (absolute paths are required):

      $ export BOSH_MICRO_CPI_RELEASE_PATH=/tmp/bosh-warden-cpi-9.tgz
      $ export BOSH_MICRO_STEMCELL_PATH=/tmp/bosh-stemcell-348-warden-boshlite-ubuntu-trusty-go_agent.tgz
      $ ./bin/test-acceptance-with-vm --provider=virtualbox

#### Running tests using AWS provider

  When using the AWS provider, you will need to provide the following:

      $ export BOSH_MICRO_PRIVATE_KEY=/tmp/id_rsa

      # The following variables are required by Bosh Lite
      $ export BOSH_AWS_ACCESS_KEY_ID=access_key
      $ export BOSH_AWS_SECRET_ACCESS_KEY=secret
      $ export BOSH_LITE_KEYPAIR=keypair
      $ export BOSH_LITE_SECURITY_GROUP=sg-1234
      $ export BOSH_LITE_PRIVATE_KEY=$BOSH_MICRO_PRIVATE_KEY

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
}
```

Run tests:

```
BOSH_MICRO_CONFIG_PATH=config.json bin/test-acceptance
```


## Using UAA for authentication

## 1. Grab a copy of the cf release.
[Go here](http://bosh.io/releases/github.com/cloudfoundry/cf-release) and wait long time for download.

## 2. Add stuff to your manifest.
Add this to your bosh cluster jobs:

    - { name: uaa, release: cf }

Here's the minimal set of properties:

```yaml
uaa:
  admin:
    client_secret: PASSWORD
  batch:
    password: PASSWORD
    username: batch_user
  clients:
    hm:
      secret: PASSWORD
    login:
      authorities: oauth.login,scim.write,clients.read,notifications.write,critical_notifications.write,emails.write,scim.userids,password.write
      authorized-grant-types: authorization_code,client_credentials,refresh_token
      redirect-uri: http://login.REPLACE_WITH_SYSTEM_DOMAIN
      scope: openid,oauth.approvals
      secret: PASSWORD
  jwt:
    signing_key: |
      -----BEGIN RSA PRIVATE KEY-----
      MIICXAIBAAKBgQDHFr+KICms+tuT1OXJwhCUmR2dKVy7psa8xzElSyzqx7oJyfJ1
      JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMXqHxf+ZH9BL1gk9Y6kCnbM5R6
      0gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBugspULZVNRxq7veq/fzwIDAQAB
      AoGBAJ8dRTQFhIllbHx4GLbpTQsWXJ6w4hZvskJKCLM/o8R4n+0W45pQ1xEiYKdA
      Z/DRcnjltylRImBD8XuLL8iYOQSZXNMb1h3g5/UGbUXLmCgQLOUUlnYt34QOQm+0
      KvUqfMSFBbKMsYBAoQmNdTHBaz3dZa8ON9hh/f5TT8u0OWNRAkEA5opzsIXv+52J
      duc1VGyX3SwlxiE2dStW8wZqGiuLH142n6MKnkLU4ctNLiclw6BZePXFZYIK+AkE
      xQ+k16je5QJBAN0TIKMPWIbbHVr5rkdUqOyezlFFWYOwnMmw/BKa1d3zp54VP/P8
      +5aQ2d4sMoKEOfdWH7UqMe3FszfYFvSu5KMCQFMYeFaaEEP7Jn8rGzfQ5HQd44ek
      lQJqmq6CE2BXbY/i34FuvPcKU70HEEygY6Y9d8J3o6zQ0K9SYNu+pcXt4lkCQA3h
      jJQQe5uEGJTExqed7jllQ0khFJzLMx0K6tj0NeeIzAaGCQz13oo2sCdeGRHO4aDh
      HH6Qlq/6UOV5wP8+GAcCQFgRCcB+hrje8hfEEefHcFpyKH+5g1Eu1k0mLrxK2zd+
      4SlotYRHgPCEubokb2S1zfZDWIXW3HmggnGgM949TlY=
      -----END RSA PRIVATE KEY-----
    verification_key: |
      -----BEGIN PUBLIC KEY-----
      MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDHFr+KICms+tuT1OXJwhCUmR2d
      KVy7psa8xzElSyzqx7oJyfJ1JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMX
      qHxf+ZH9BL1gk9Y6kCnbM5R60gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBug
      spULZVNRxq7veq/fzwIDAQAB
      -----END PUBLIC KEY-----
  scim:
    users:
    - admin|PASSWORD|scim.write,scim.read
  url: http://uaa.example.com
  login:
    client_secret: PASSWORD
domain: example.com
nats:
  password: nats
  port: 4222
  user: nats
  machines: []
networks:
  apps: default
uaadb:
  address: 10.0.16.101
  databases:
  - name: bosh
    tag: uaa
  db_scheme: postgresql
  port: 5524
  roles:
  - name: postgres
    password: postgres
    tag: admin
login:
  protocol: http
```
