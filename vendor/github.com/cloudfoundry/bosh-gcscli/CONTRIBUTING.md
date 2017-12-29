# Contributing

## Submitting an Issue
We use the [GitHub issue tracker](https://github.com/cloudfoundry/bosh-gcscli/issues) to track bugs and features.
Before submitting a bug report or feature request, check to make sure it hasn't already been submitted. You can indicate
support for an existing issue by voting it up. When submitting a bug report, please include a
[Gist](http://gist.github.com/) that includes a stack trace and any details that may be necessary to reproduce the bug,
including your gem version, Ruby version, and operating system. Ideally, a bug report should include a pull request with failing specs.

## Submitting a Pull Request
You can add a feature or bug-fix via pull request.
1. Fork the project
1. Create a branch for your feature or fix from the `develop` branch. Replace `your-feature-name` with a description of your feature or fix:
   ```
   git checkout -b your-feature-name develop
   ```
1. Implement your feature or bug fix
1. Commit and push your changes
1. Submit a pull request to the `develop` branch of the [bosh-gcscli] repository. PRs to the master branch are not accepted.
1. Unit tests and a BOSH release are created for each PR. You should see the status of your PR change to "pending" within a few minutes of submitting it, and then to "passed" or "failed" within 10 minutes.

[bosh-gcscli]: https://github.com/cloudfoundry/bosh-gcscli/