# Create deployment manifest
This file will be used by the BOSH Micro CLI to deploy Micro BOSH.

### Example deployment manifest

```yaml
---
name: micro-bosh

networks:
- name: default
  type: dynamic
  cloud_properties:
    subnet: AWS_SUBNET_NAME
- name: vip
  type: vip

resource_pools:
- name: default
  cloud_properties:
    instance_type: AWS_INSTANCE_TYPE
    availability_zone: AWS_AVAILABILITY_ZONE

cloud_provider:
  ssh_tunnel:
    host: MICRO_BOSH_IP
    port: 22
    user: ssh-user
    password: ssh-password
  registry:
    username: registry-user
    password: registry-password
    port: 6901
    host: localhost
  mbus: https://mbus-user:mbus-password@MICRO_BOSH_IP:6868
  properties: # properties that are saved in registry by CPI for the agent
    blobstore:
      provider: local
      path: /var/vcap/micro_bosh/data/cache
    registry:
      username: admin
      password: admin
      port: 6901
      host: localhost
    ntp: [NTP_ADDRESS]
    aws:
      access_key_id: AWS_SECRET_KEY
      secret_access_key: AWS_ACCESS_KEY
      default_key_name: AWS_KEY_NAME
      default_security_groups: [AWS_SECURITY_GROUP_NAME]
      region: AWS_REGION
      ec2_private_key: PATH_TO_PRIVATE_KEY
    agent:
      mbus: https://mbus-user:mbus-password@0.0.0.0:6868

jobs:
- name: bosh
  templates:
  - name: nats
  - name: redis
  - name: postgres
  - name: powerdns
  - name: blobstore
  - name: director
  - name: health_monitor
  - name: registry
  networks:
  - name: default
  - name: vip
    static_ips:
    - MICRO_BOSH_IP
  properties: # properties that are used to render job templates
    nats:
      user: nats
      password: nats
      address: 127.0.0.1
    ...

```
See [https://github.com/cloudfoundry/bosh/tree/master/release/jobs](https://github.com/cloudfoundry/bosh/tree/master/release/jobs) for defaults

# Set deployment manifest

The command below sets the deployment manifest. The current deployment path is saved to `~/.bosh_micro.json`.

    bosh-micro deployment manifest.yml

# Deploy Microbosh

The command below deploys Micro BOSH with the provided CPI release and stemcell.

    bosh-micro deploy cpi-release.tgz stemcell.tgz

Once the deploy is finished, Micro BOSH will be available to be targeted.

---
# Deployment Flow
This section describes how the CLI works. These steps are performed by the CLI.

For additional information see the [decision tree](micro-cli-flow.png) of the deploy command.

## 1. Validating manifest, release and stemcell

The first step of the deploy process is validation. As part of that validation the CLI verifies if there are changes in either manifest, release or stemcell. In case there are no changes CLI will exit early with message `Skipping deploy`.

As part of manifest validation BOSH Micro CLI validates manifest properties and parses manifest for deploy. BOSH Micro CLI parses the deployment manifest into two parts: the BOSH deployment manifest, and the CPI deployment manifest.

The BOSH deployment manifest is used to deploy Micro BOSH. The Micro BOSH is defined by the `networks`, `resource_pools`, and `jobs` sections of the manifest. The BOSH micro job must be defined as the first job in the `jobs` section. Any other job will be ignored.

The CPI deployment manifest is used to deploy the CPI locally. It is constructed from the `cloud_provider` section of the manifest.

## 2. Deploying CPI Release

The provided CPI release is compiled on the machine where `bosh-micro deploy` is run, and is used locally to run the CPI commands necessary to create the Micro BOSH.

The CPI release must contain a job called `cpi`. During CPI release deployment, all the packages that the `cpi` job depends on will be compiled and their templates rendered. CPI job templates have access to properties defined in the `cloud_provider -> properties` section of the manifest.

The compiled packages and rendered job templates are stored in a `~/.bosh_micro/<deployment_uuid>` folder for each deployment.

## 3. Uploading Stemcell

After the CPI is deployed locally, the CLI calls the `create_stemcell` CPI method with the provided stemcell.

## 4. Starting Registry

Before deploying Micro BOSH, the CLI starts the registry. The registry can be used by the CPI to store mutable data to be later accessed by the agent on the Micro BOSH VM. The registry is a service to store mutable data when the infrastructure's metadata service is immutable. This data is anything that is not known until after the CPI creates the VM that the agent will require. For example, information about any persistent disks that are attached to Micro BOSH after the Micro BOSH VM is created can be stored in the registry.

The CPI will store the registry URL in the infrastructure's metadata service. The agent on the Micro BOSH VM will fetch registry settings from the provided URL.

## 5. Deleting existing VM

In case the VM was previosly deployed, the CLI tries to connect to the agent on the existing VM. If the agent is responsive, the CLI stops services that are running on that VM and unmounts all disks that are attached to the VM. Eventually, the CLI deletes the existing VM and removes VM CID from `deployment.json`.

## 6. Creating new VM

Next, the CLI sends the `create_vm` command to the CPI with the properties parsed from the manifest. Additionally, the VM CID is persisted in `deployment.json` in the same folder as the deployment manifest.

## 7. Starting SSH Tunnel

The CLI creates a reverse SSH tunnel to Micro BOSH VM using the properties provided in the manifest. This allows the agent on the Micro BOSH VM to access the registry, which is running on the machine where `bosh-micro deploy` was run.

## 8. Waiting for Agent

Once the SSH tunnel is up the CLI uses provided mbus URL to issue ping messages to the agent on the Micro BOSH VM. Once the agent is ready it will respond to the ping.

## 9. Sending stop message

Once agent is listening on Mbus endpoint micro CLI sends stop message to the agent. The agent is using `monit` to manage job states on VM. The stop is a preparation for the subsequent job update.

## 10. Sending micro BOSH apply spec

Next micro CLI sends apply message with the list of packages and jobs that should be installed on VM. The agent serves a blobstore at `<Mbus URL>/blobs` endpoint. The package and job list is parsed from `apply_spec.yml` which is included in the stemcell.

For each of the template specified in micro BOSH job micro CLI downloads corresponding job template from the blobstore, renders the template with the properties specified for micro BOSH job in deployment manifest. Once all the templates are rendered micro CLI uploads the archive of all the rendered templates to the blobstore and generates an apply message. Apply message contains the list of all packages, spec of templates archive with uploaded blob ID, networks spec parsed from deployment manifest and configuration hash which is a digest of all rendered job template files.

## 11. Sending start message

Once `apply` task is finished micro CLI sends `start` message to the agent which starts installed jobs.

## 12. Creating disk

ClI will create and attach a disk to Micro BOSH VM if it is requested in manifest. There are two ways to request the disk:

1. Adding `persistent_disk_pool` property on a Micro BOSH job which references the disk pool in the list of `disk_pools` specified on the top level of the manifest.
2. Adding `persistent_disk` property which specifies the size of persistent disk.

You should use `disk_pools` if you want to use disk cloud_properties.

In this case the CLI calls the `create_disk` CPI method with the provided size. Additionally, the disk CID is persisted in `deployment.json` in the same folder as the deployment manifest.

## 13. Attaching disk

After disk is created CLI calls `attach_disk` CPI method. After disk is attached CLI issues `mount_disk` request to the agent on the Micro BOSH VM.

# To be continued…
