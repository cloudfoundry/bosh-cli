## bosh-gcscli

[![GoDoc](https://godoc.org/github.com/cloudfoundry/bosh-gcscli?status.svg)](https://godoc.org/github.com/cloudfoundry/bosh-gcscli)


A Golang CLI for uploading, fetching and deleting content to/from [Google Cloud Storage](https://cloud.google.com/storage/). 
This tool exists to work with the [bosh-cli](https://github.com/cloudfoundry/bosh-cli) and [director](https://github.com/cloudfoundry/bosh).

This is **not** an official Google Product.

## Installation

```bash
go get github.com/cloudfoundry/bosh-gcscli
```

## Commands

### Usage
```bash
bosh-gcscli --help
```
### Upload an object
```bash
bosh-gcscli -c config.json put <path/to/file> <remote-blob>
```
### Fetch an object
```bash
bosh-gcscli -c config.json get <remote-blob> <path/to/file>
```
### Delete an object
```bash
bosh-gcscli -c config.json delete <remote-blob>
```
### Check if an object exists
```bash
bosh-gcscli -c config.json exists <remote-blob>```
```

## Configuration
The command line tool expects a JSON configuration file. Run `bosh-gcscli --help` for details.

### Authentication Methods (`credentials_source`)
* `static`: A [service account](https://cloud.google.com/iam/docs/creating-managing-service-account-keys) key will be provided via the `json_key` field.
* `none`: No credentials are provided. The client is reading from a public bucket.
* &lt;empty&gt;: [Application Default Credentials](https://developers.google.com/identity/protocols/application-default-credentials)
  will be used if they exist (either through `gcloud auth application-default login` or a [service account](https://cloud.google.com/iam/docs/understanding-service-accounts)).
  If they don't exist the client will fall back to `none` behavior.

## Running Integration Tests

1. Ensure [gcloud](https://cloud.google.com/sdk/downloads) is installed and you have authenticated (`gcloud auth login`).
   These credentials will be used by the Makefile to create/destroy Google Cloud Storage buckets for testing.
1. Set the Google Cloud project: `gcloud config set project <your project>`
1. Generate a service account with the `Storage Admin` role for your project and set the contents as 
    the environment variable `GOOGLE_APPLICATION_CREDENTIALS`, for example:
   ```bash
   export project_id=$(gcloud config get-value project)

   export service_account_name=bosh-gcscli-integration-tests
   export service_account_email=${service_account_name}@${project_id}.iam.gserviceaccount.com
   credentials_file=$(mktemp)

   gcloud config set project ${project_id}
   gcloud iam service-accounts create ${service_account_name} --display-name "Integration Test Access for bosh-gcscli"
   gcloud iam service-accounts keys create ${credentials_file} --iam-account ${service_account_email}
   gcloud project add-iam-policy-binding ${project_id} --member serviceAccount:${service_account_email} --role roles/storage.admin
  
   export GOOGLE_APPLICATION_CREDENTIALS="$(cat ${credentials_file})"
   ```
1. Run the unit and fast integration tests: `make test-fast-int`
1. Clean up buckets: `make clean-gcs`

## Development

* A Makefile is provided that automates integration testing. Try `make help` to get started.
* [gvt](https://godoc.org/github.com/FiloSottile/gvt) is used for vendoring.

## Contributing

For details on how to contribute to this project - including filing bug reports and contributing code changes - please see [CONTRIBUTING.md](./CONTRIBUTING.md).

## License

This tool is licensed under Apache 2.0. Full license text is available in [LICENSE](LICENSE).