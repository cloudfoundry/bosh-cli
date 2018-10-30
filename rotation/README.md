# Certificate Rotation

Many `bosh` deployments use generated certificates for authentication between components.

While `bosh` and `credhub` can generate these automatically, they expire after 1 year, which is problematic.

We have built an experimental flag to rotate certificates in a deployment for you.

## Pre-requisites

1. `bosh` must be configured to use `credhub` as a configuration server. It should be easy to adapt this technique for other configuration options, but that's all that is supported so far.

2. Before running with this option, you need to ensure the following environment variables are set so that the `bosh` CLI can communicate directly with `credhub`. e.g. you might set as follows (assuming a `creds.yml` for your `bosh` is nearby):

    ```bash
    export CREDHUB_SERVER="https://director.bosh.cld.internal:8844"
    export CREDHUB_CLIENT=credhub-admin
    export CREDHUB_SECRET="$(bosh int creds.yml --path /credhub_admin_client_secret)"
    export CREDHUB_CA_CERT="$(cat <<EOF
    $(bosh int creds.yml --path /credhub_tls/ca)
    $(bosh int creds.yml --path /uaa_ssl/ca)
    EOF
    )"
    ```

3. This has only been tested with `cf-deployment` version `5.4.0`. There is at least one change in that particular version that is known to help this process succeed, so best not to try with older versions.

## Build and install

Currently this exists as a fork of bosh-cli v5.3.1. To build:

```bash
go get github.com/cloudfoundry/bosh-cli
cd ${GOPATH:-$HOME/go}/src/github.com/cloudfoundry/bosh-cli
git remote add govau https://github.com/govau/bosh-cli
git fetch govau
git checkout govau
go install github.com/cloudfoundry/bosh-cli

# copy binary to somewhere on your path
cp ${GOPATH:-$HOME/go}/bin/bosh-cli /path/to/where/you/normally/put/bosh
```

## Running

This flag should be used during a deploy, e.g.:

```bash
bosh -d cf deploy cf-deployment/cf-deployment.yml \
  -v system_domain=example.com \
  --no-redact \
  -n \
  ...
  --progressive-certificate-rotation
```

When this executes, the first thing that it will do is look for the value `credential_rotation_action` in `credhub` for the deployment.

If this value is not set, then no special actions will take place and deployment will occur normally.

To initiate certificate rotation, run the following *before* invoking `bosh deploy` (use the correct bosh director and deployment name for your installation):

```bash
credhub set -n /main/cf/credential_rotation_action -t value -v create-and-deploy-transitional-cas
```

And then, when calling `bosh deploy ... --progressive-certificate-rotation`, you should see the following output:

```
Using deployment 'cf'

Beginning certificate rotation process, step 1 of 3 - Creating transitional CAs...

Creating new transitional CA: /main/cf/service_cf_internal_ca
Creating new transitional CA: /main/cf/application_ca
Creating new transitional CA: /main/cf/silk_ca
Creating new transitional CA: /main/cf/log_cache_ca
Creating new transitional CA: /main/cf/uaa_ca
Creating new transitional CA: /main/cf/consul_agent_ca
Creating new transitional CA: /main/cf/network_policy_ca
Creating new transitional CA: /main/cf/diego_instance_identity_ca
Creating new transitional CA: /main/cf/credhub_ca
Creating new transitional CA: /main/cf/loggregator_ca

... a normal deployment occurs ...

Continuing certificate rotation process, step 2 of 3 - regenerating leaf certificates...

Deleting leaf certificate: /main/cf/cc_tls
Deleting leaf certificate: /main/cf/log_cache
Deleting leaf certificate: /main/cf/scheduler_api_tls
Deleting leaf certificate: /main/cf/adapter_rlp_tls
Deleting leaf certificate: /main/cf/logs_provider
Deleting leaf certificate: /main/cf/loggregator_tls_rlp
Deleting leaf certificate: /main/cf/gorouter_backend_tls
Deleting leaf certificate: /main/cf/uaa_login_saml
Deleting leaf certificate: /main/cf/adapter_tls
Deleting leaf certificate: /main/cf/loggregator_rlp_gateway
Deleting leaf certificate: /main/cf/network_policy_server
Deleting leaf certificate: /main/cf/cc_bridge_cc_uploader_server
Deleting leaf certificate: /main/cf/cc_bridge_tps
Deleting leaf certificate: /main/cf/log_cache_tls_cc_auth_proxy
Deleting leaf certificate: /main/cf/scheduler_client_tls
Deleting leaf certificate: /main/cf/loggregator_tls_doppler
Deleting leaf certificate: /main/cf/silk_controller
Deleting leaf certificate: /main/cf/credhub_tls
Deleting leaf certificate: /main/cf/diego_locket_client
Deleting leaf certificate: /main/cf/cc_public_tls
Deleting leaf certificate: /main/cf/loggregator_rlp_gateway_tls_cc
Deleting leaf certificate: /main/cf/loggregator_tls_agent
Deleting leaf certificate: /main/cf/diego_rep_client
Deleting leaf certificate: /main/cf/silk_daemon
Deleting leaf certificate: /main/cf/diego_auctioneer_client
Deleting leaf certificate: /main/cf/diego_locket_server
Deleting leaf certificate: /main/cf/uaa_ssl
Deleting leaf certificate: /main/cf/loggregator_tls_cc_tc
Deleting leaf certificate: /main/cf/loggregator_tls_tc
Deleting leaf certificate: /main/cf/loggregator_tls_statsdinjector
Deleting leaf certificate: /main/cf/diego_bbs_server
Deleting leaf certificate: /main/cf/diego_bbs_client
Deleting leaf certificate: /main/cf/network_policy_client
Deleting leaf certificate: /main/cf/cc_bridge_cc_uploader
Deleting leaf certificate: /main/cf/diego_rep_agent_v2
Deleting leaf certificate: /main/cf/diego_auctioneer_server

Flipping transitional flags for: /main/cf/network_policy_ca
Flipping transitional flags for: /main/cf/consul_agent_ca
Flipping transitional flags for: /main/cf/log_cache_ca
Flipping transitional flags for: /main/cf/silk_ca
Flipping transitional flags for: /main/cf/application_ca
Flipping transitional flags for: /main/cf/tls_credhub_ca
Flipping transitional flags for: /main/cf/loggregator_ca
Flipping transitional flags for: /main/cf/credhub_ca
Flipping transitional flags for: /main/cf/diego_instance_identity_ca
Flipping transitional flags for: /main/cf/uaa_ca
Flipping transitional flags for: /main/cf/service_cf_internal_ca

... a second deployment occurs ...

Continuing certificate rotation process, step 3 of 3 - removing legacy CAs...

Removing transitional flags for: /main/cf/loggregator_ca
Removing transitional flags for: /main/cf/credhub_ca
Removing transitional flags for: /main/cf/log_cache_ca
Removing transitional flags for: /main/cf/silk_ca
Removing transitional flags for: /main/cf/service_cf_internal_ca
Removing transitional flags for: /main/cf/diego_instance_identity_ca
Removing transitional flags for: /main/cf/network_policy_ca
Removing transitional flags for: /main/cf/consul_agent_ca
Removing transitional flags for: /main/cf/uaa_ca
Removing transitional flags for: /main/cf/application_ca

... a third deployment occurs ...
```

After each successful deployment, the `credential_rotation_action` value is updated so that if any of the deployments fails for any reason, and then same command is run again, it will skip the deployments that previously succeeded and should be safe to re-run.

After the third and final deployment, `credential_rotation_action` is set to `no-action-needed`, which means that if the command is re-run, no certificate rotation actions are taken.

## How to check certificate lifetimes?

This bash snippet will enumerate certificates from your `credhub` and print how many days remaining:

```bash
# requires: credhub, jq, openssl
now="$(date "+%s")"
names="$(credhub find --output-json | jq -r .credentials[].name)"
for name in $names; do
    cert="$(credhub get --output-json -n "${name}" | jq -r 'select(.type=="certificate")|.value.certificate')"
    if [[ $cert ]]; then
        notafter="$(openssl x509 -noout -enddate -in <(echo "${cert}") | cut -c 10-)"
        certunix="$(date -d "${notafter}" "+%s")"
        days="$(expr \( ${certunix} - ${now} \) / 86400)"
        echo "${name} will expires in ${days} days"
    fi
done
```

Output:

```
/main/cf/credhub_tls will expires in 364 days
/main/cf/gorouter_backend_tls will expires in 364 days

...

/dns_healthcheck_server_tls will expires in 274 days
/dns_healthcheck_tls_ca will expires in 274 days
```

Note this won't show certificates that are used by `bosh` itself, in `creds.yml`, but this will:

```bash
# requires: yq, openssl
creds_yml_path="/path/to/bosh/creds.yml"
names="$(cat "$creds_yml_path" | yq keys | jq -r .[])"
now="$(date "+%s")"
for name in $names; do
    hascert="$((cat "$creds_yml_path" | yq ".${name} | has(\"certificate\")" 2> /dev/null) || echo "false")"
    if [ "$hascert" = "true" ]; then
        cert="$(cat "$creds_yml_path" | yq -r .${name}.certificate)"
        if [[ $cert ]]; then
            notafter="$(openssl x509 -noout -enddate -in <(echo "${cert}") | cut -c 10-)"
            certunix="$(date -d "${notafter}" "+%s")"
            days="$(expr \( ${certunix} - ${now} \) / 86400)"
            echo "${name} will expires in ${days} days"
        fi
    fi
done
```

Output:

```
blobstore_ca will expires in 321 days
...
uaa_ssl will expires in 321 days
```

## Details

### Step 1: Creating transitional CAs

This enumerates all certificates from credhub that were generated for this deployment and filters this to CAs only.

`credhub` is then requested to generate a new transitional certificate for each CA.

Since the bosh director currently appears unable to fetch multiple certificates (active and transitional) from `credhub` directly, we then inject the older (but still active) CA certificates directly into the manifest, just after each of the normal equivalent variable expansions. When the bosh director receives the manifest, it then augments these with the newer transitional CA certificates so that each configured CA certificate variable expansion ends up with 2 PEM encoded certificates.

The deployment then proceeds - however at this time no new credentials are generated, rather the outcome is that each job now trusts both the old and the new CAs.

### Step 2: Regenerating leaf certificates

This first deletes all non-CA certificates in the deployment from `credhub`. When deployed, `credhub` will regenerate these with the new CA certificates using normal process.

Next the transitional flag is flipped on all CA certificates so that `credhub` will use the new CAs to generate the new leaf certificates at deployment time.

As in the previous step, the manifest is augmented in the same manner so that both old and new CAs are trusted by all jobs.

When this deployment is complete, all new credentials have been generated and are in use.

### Step 3: Removing legacy CAs

As a final step we remove the transitional flag from the old CAs, and then re-deploy without augmenting the manifest.

After this, the jobs are now configured to only trust the new CAs.

## What else needs to be done?

We note the following TODOs:

1. There are some global certificates, such as `/dns_api_tls_ca` that are not contemplated in the above.
2. There are `bosh` credentials, stored in `creds.yml` (such as `credhub_tls`) that are not contemplated above.
3. Why does `cf-deployment` need so many CAs?

## Authors

Adam Eijdenberg (cloud.gov.au) <adam.eijdenberg@digital.gov.au>
