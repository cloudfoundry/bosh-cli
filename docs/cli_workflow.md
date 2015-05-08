# Create deployment manifest

This file will be used by the bosh-init to deploy BOSH.

### Example deployment manifest

```yaml
---
name: redis

releases:
- name: bosh-aws-cpi
  url: https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=7
  sha1: 6545812c1c8245331b8c420f886dafd24b866eed
- name: redis
  url: https://bosh.io/d/github.com/cloudfoundry-community/redis-boshrelease?v=9.1
  sha1: e18fac6f755c9d8cd90d2f9fad40a7023d1c672f

resource_pools:
- name: default
  network: default
  stemcell:
    url: file://./light-bosh-stemcell-2941-aws-xen-ubuntu-trusty-go_agent.tgz
  cloud_properties:
    instance_type: m3.medium
    availability_zone: ap-northeast-1c

networks:
- name: default
  type: dynamic
  cloud_properties: {subnet: subnet-5907c031}
- name: vip
  type: vip

jobs:
- name: redis
  instances: 1
  templates:
  - {name: redis, release: redis}
  resource_pool: default
  persistent_disk: 10240
  networks:
  - {name: vip, static_ips: [52.68.164.131]}
  - name: default
  properties:
    redis: {port: 6379}

cloud_provider:
  template: {name: cpi, release: bosh-aws-cpi}

  ssh_tunnel:
    host: 52.68.164.131
    port: 22
    user: vcap
    private_key: ./bosh.pem

  mbus: https://nats:nats@52.68.164.131:6868

  properties:
    aws:
      access_key_id: AKI...
      secret_access_key: 0kw...
      default_key_name: bosh
      default_security_groups: [bosh]
      region: ap-northeast-1

    agent: {mbus: "https://nats:nats@0.0.0.0:6868"}

    blobstore: {provider: local, path: /var/vcap/micro_bosh/data/cache}

    ntp: [0.north-america.pool.ntp.org]
```

See [https://github.com/cloudfoundry/bosh/tree/master/release/jobs](https://github.com/cloudfoundry/bosh/tree/master/release/jobs) for defaults

# Deploy VM

The command below deploys VM with given releases using CPI release and stemcell.

```
bosh-init deploy redis.yml
```

---

# Deployment Flow

This section describes how the CLI works. These steps are performed by the CLI.

For additional information see the [decision tree](init-cli-flow.png) of the deploy command.

## 1. Validating manifest, release and stemcell

The first step of the deploy process is validation. As part of that validation the CLI verifies if there are changes in either manifest, release or stemcell. In case there are no changes CLI will exit early with message `Skipping deploy`.

As part of manifest validation the CLI validates manifest properties and parses manifest for deploy. The CLI parses the deployment manifest into two parts: the deployment manifest, and the CPI configuration.

The deployment manifest is used arbitrary releases onto a single VM. The deployment manifest is defined by the `networks`, `resource_pools`, `disk_pools`, and `jobs` sections of the manifest. Currently only one job is allowed to be specified since the CLI will only create single VM.

The CPI configuration is used to install and configure the CPI locally. It is constructed from the `cloud_provider` section of the manifest.

## 2. Installing CPI Release

The provided CPI release is compiled on the machine where `bosh-init` is run, and is used locally to run the CPI commands necessary to create the VM.

The CPI release must contain a job specified by the `cloud_provider.template.job`. During CPI installation, all the packages that the CPI job depends on will be compiled and their templates rendered. CPI job templates have access to properties defined in the `cloud_provider -> properties` section of the manifest.

The compiled packages and rendered job templates are stored in a `~/.bosh_init/<installation_id>` folder for each deployment.

## 3. Uploading Stemcell

After the CPI is installed locally, the CLI calls the `create_stemcell` CPI method with the provided stemcell.

## 4. Starting Registry

Before creating a VM, the CLI starts the registry. The registry can be used by the CPI to store mutable data to be later accessed by the agent running on the VM. The registry is a service to store mutable data when the infrastructure's metadata service is immutable. This data is anything that is not known until after the CPI creates the VM that the agent will require. For example, information about any persistent disks that are attached to Micro BOSH after the Micro BOSH VM is created can be stored in the registry.

The CPI will store the registry URL in the infrastructure's metadata service. The agent on the VM will fetch registry settings from the provided URL.

Note: We are planning to eventually remove the registry to simplify how CPIs behave.

## 5. Deleting existing VM

In case the VM was previosly deployed, the CLI tries to connect to the agent on the existing VM. If the agent is responsive, the CLI stops services that are running on that VM and unmounts all disks that are attached to the VM. Eventually, the CLI deletes the existing VM and removes VM CID from deployment state file.

## 6. Creating new VM

Next, the CLI sends the `create_vm` command to the CPI with the properties parsed from the manifest. Additionally, the VM CID is persisted in deployment state file in the same folder as the deployment manifest.

## 7. Starting SSH Tunnel

The CLI creates a reverse SSH tunnel to BOSH VM using the properties provided in the manifest. This allows the agent on the VM to access the registry, which is running on the machine where `bosh-init deploy` was run.

## 8. Waiting for Agent

Once the SSH tunnel is up the CLI uses provided mbus URL to issue ping messages to the agent on the Micro BOSH VM. Once the agent is ready it will respond to the ping.

## 9. Creating disk

The CLI will create and attach a disk to VM if it is requested in the deployment manifest. There are two ways to request the disk:

1. Adding `persistent_disk_pool` property on a job which references the disk pool in the list of `disk_pools` specified on the top level of the manifest.
2. Adding `persistent_disk` property which specifies the size of persistent disk.

You should use `disk_pools` if you want to use disk `cloud_properties`.

In this case the CLI calls the `create_disk` CPI method with the provided size. Additionally, the disk CID is persisted in deployment state file.

## 10. Attaching disk

After disk is created CLI calls `attach_disk` CPI method. After disk is attached CLI issues `mount_disk` request to the agent on the Micro BOSH VM.

## 11. Sending stop message

Once agent is listening on mbus URL, the CLI sends stop message to the agent. The agent is using `monit` to manage job states on VM. The stop is a preparation for the subsequent job update.

## 12. Sending apply message

Next the CLI sends apply message with the list of packages and jobs that should be installed on VM. The agent serves a blobstore at `<mbus URL>/blobs` endpoint.

For each of the template specified, the CLI downloads corresponding job template from the blobstore, renders the template with the properties specified for job in deployment manifest. Once all the templates are rendered the CLI uploads the archive of all the rendered templates to the blobstore and generates an apply message. Apply message contains the list of all packages, spec of templates archive with uploaded blob ID, networks spec parsed from deployment manifest and configuration hash which is a digest of all rendered job template files.

## 13. Sending start message

Once `apply` task is finished the CLI sends `start` message to the agent which starts installed jobs.
