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

1. Tell bosh-micro which deployment manifest to use:

    ```
    out/bosh-micro deployment manifest.yml
    ```

      See the [CLI workflow](docs/cli_workflow.md) for more information on creating a manifest.

1. Deploy

    ```
    out/bosh-micro deploy stemcell.tgz cpi-release.tgz bosh-release.tgz
    ```

    Where:
  
    - `stemcell.tgz` is a BOSH stemcell appropriate for the CPI release
    - `cpi-release.tgz` is a BOSH CPI release
    - `bosh-release.tgz` is a BOSH release of BOSH    


Once deployed, the BOSH director can be targeted using the [bosh_cli](https://rubygems.org/gems/bosh_cli)].

## Example Deploy Output

The following output (printed to STDOUT) is from deploying BOSH into a Warden Container (inside a bosh-lite vagrant VM).

```
> bosh-micro deploy /home/vagrant/test-warden-stemcell.tgz /home/vagrant/bosh-warden-cpi-16.tgz /home/vagrant/bosh-2811.tgz
Deployment manifest: '/home/vagrant/manifest.yml'
Deployment state: '/home/vagrant/deployment.json'

Started validating
  Validating stemcell... Finished (00:00:04)
  Validating releases... Finished (00:00:03)
  Validating deployment manifest... Finished (00:00:00)
  Validating cpi release... Finished (00:00:00)
Finished validating (00:00:07)

Started installing CPI
  Compiling package 'golang_1.3/fc3bc1b4431e8913d91362c1183c9852809d35f6'... Finished (00:00:10)
  Compiling package 'cpi/6f5b7e1d1050764cd14da9cc8e8683a03a502996'... Finished (00:00:04)
  Rendering job templates... Finished (00:00:00)
  Installing packages... Finished (00:00:01)
  Installing job 'cpi'... Finished (00:00:00)
Finished installing CPI (00:00:16)

Starting registry... Finished (00:00:00)
Uploading stemcell 'bosh-warden-boshlite-ubuntu-trusty-go_agent/0000'... Finished (00:00:14)

Started deploying
  Creating VM for instance 'bosh/0' from stemcell '47017a4e-4a81-41cf-4afc-1121346d46b4'... Finished (00:00:00)
  Waiting for the agent on VM '1987aaea-eb8a-4905-54d3-88202ce550d4' to be ready... Finished (00:00:01)
  Creating disk... Finished (00:00:00)
  Attaching disk '030015fc-4148-4313-5e17-608dc4b7aa76' to VM '1987aaea-eb8a-4905-54d3-88202ce550d4'... Finished (00:00:01)
  Compiling package 'ruby/8c1c0bba2f15f89e3129213e3877dd40e339592f'... Finished (00:01:32)
  Compiling package 'postgres/aa7f5b110e8b368eeb8f5dd032e1cab66d8614ce'... Finished (00:00:04)
  Compiling package 'nginx/8f131f14088764682ebd9ff399707f8adb9a5038'... Finished (00:00:33)
  Compiling package 'libpq/6aa19afb153dc276924693dc724760664ce61593'... Finished (00:00:14)
  Compiling package 'mysql/e5309aed88f5cc662bc77988a31874461f7c4fb8'... Finished (00:00:06)
  Compiling package 'redis/ec27a0b7849863bc160ac54ce667ecacd07fc4cb'... Finished (00:00:24)
  Compiling package 'powerdns/e41baf8e236b5fed52ba3c33cf646e4b2e0d5a4e'... Finished (00:00:01)
  Compiling package 'genisoimage/008d332ba1471bccf9d9aeb64c258fdd4bf76201'... Finished (00:00:16)
  Compiling package 'director/a59aa6cf382b0c6df4206219f9f661b86dfc6103'... Finished (00:00:37)
  Compiling package 'nats/6a31c7bb0d5ffa2a9f43c7fd7193193438e20e92'... Finished (00:00:09)
  Compiling package 'health_monitor/a8a4a1cb04f924f17f9944845f5f4a73ecd4b895'... Finished (00:00:18)
  Rendering job templates... Finished (00:00:00)
  Updating instance 'bosh/0'... Finished (00:00:09)
  Waiting for instance 'bosh/0' to be running... Finished (00:00:07)
Finished deploying (00:04:37)
```

## Logging

Along with the UI output (STDOUT) and UI errors (STDERR), bosh-micro can output more verbose logs.

Logging is disabled by default (`BOSH_MICRO_LOG_LEVEL` defaults to NONE).

To enable logging, set the `BOSH_MICRO_LOG_LEVEL` environment variable to one of the following values:

DEBUG, INFO, WARN, ERROR, NONE (default)

Logs write to STDOUT (debug & info) & STDERR (warn & error) by default.

To write logs to a file, set the `BOSH_MICRO_LOG_PATH` environment variable to the path of the file to create and/or append to.

## Deployment State

The current state of your deployment is stored in a `deployment.json` file in the same directory as your deployment manifest.

This allows you to deploy multiple deployments with different manifests, as long as they're in different directories.

Do not delete this file unless you have already deleted your deployment (with `bosh-micro delete` or by manually removing the VM, disk(s), & stemcell from the infrastructure).


## Other

To build bosh-micro see our [workstation setup docs](https://github.com/cloudfoundry/bosh-micro-cli/blob/master/docs/build.md).

To run bosh-micro tests see our [test docs](https://github.com/cloudfoundry/bosh-micro-cli/blob/master/docs/test.md).

To deploy BOSH with UAA using bosh-micro see our [UAA docs](https://github.com/cloudfoundry/bosh-micro-cli/blob/master/docs/uaa.md).

To learn more about the bosh-micro design see our [architecture docs](https://github.com/cloudfoundry/bosh-micro-cli/blob/master/docs/architecture.md).
