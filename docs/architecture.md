## Architecture

### Deploy Command Flow

The deploy command consumes:

1. combo manifest (installation & deployment manifests),
1. stemcell (root file system),
1. CPI release, and
1. BOSH release,

The deploy command produces:

1. a local installation of the CPI
1. a remote deployment of BOSH (and its multiple jobs) on a single VM or container on the cloud infrastructure targeted by the CPI

![bosh-micro deploy flow](https://github.com/cloudfoundry/bosh-micro-cli/blob/master/docs/bosh-micro-deploy-flow.png "bosh-micro deploy flow")
