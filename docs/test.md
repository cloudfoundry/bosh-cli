## Unit Tests

Each package in the CLI has its own unit tests and there are integration tests in the `integration` package.

You can also run all tests with `bin/test`.

## Acceptance Tests

The acceptance tests are designed to exercise the main commands of the CLI (deployment, deploy, delete).

They are not designed to verify the compatibility of CPIs or testing BOSH releases.

The acceptance test related to compiled releases uses an already compiled release that was compiled against a ubuntu-trusty/2776 stemcell. As of this writing create-env does not validate that compiled releases match the stemcell. If this validation is ever added the release will need to be recompiled.

### Fly executing the acceptance tests

The task has no `image_resource` — the job supplies the image. Base the one-off
build on the pipeline's `test-acceptance` job to pick up its image, CPI release
and stemcell, and override only the source under test:

```bash
fly -t bosh execute -p \
  -c ci/tasks/test-acceptance.yml \
  --inputs-from bosh-cli/test-acceptance \
  --image docker-cpi-image \
  -i bosh-cli=<path-to-source-dir>
```

`-p` is required because the task brings up its own docker daemon by sourcing `start-docker`.
