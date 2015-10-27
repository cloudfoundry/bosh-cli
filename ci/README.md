# Steps for Configuring the BOSH Pipeline

- [ ] 0. Set up environment
      ``` bash
      export PROJECT_PATH=</PATH/TO/PROJECT>
      export LASTPASS_USER=<USERNAME@pivotal.io>
      export LASTPASS_NOTE="bosh-init concourse secrets"
      ```
- [ ] 1. Configure the pipeline
      ``` bash
      # Update project
      cd $PROJECT_PATH
      git co develop
      git pull

      # Get pipeline secrets (see "lpass" installation notes below)
      lpass login $LASTPASS_USER
      lpass show --notes "${LASTPASS_NOTE}" > /tmp/bosh-init-secrets.yml

      # Configure the pipeline
      fly -t production configure -c ci/pipeline.yml \
        --vf /tmp/bosh-init-secrets.yml bosh-init

      # Clean up secrets
      rm /tmp/bosh-init-secrets.yml
      ```

## Notes

- To install the LastPass CLI:
  ``` bash
  brew install lastpass-cli --with-pinentry
  ```
