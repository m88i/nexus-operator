# Contributing to Nexus Operator

## Have found a bug or have a feature request?

Please, open an issue for us. That will really help us improving the Operator and it would benefit other users such as yourself. There are templates ready for bug reporting and feature request to make things easier for you.

## Have any questions?

We're happy to answer! Either [open an issue](https://github.com/m88i/nexus-operator/issues) or send an email to our mailing list: [nexus-operator@googlegroups.com](mailto:nexus-operator@googlegroups.com).

## Are you willing to send a PR?

Before sending a PR, consider opening an issue first. This way we can discuss your approach and the motivations behind it while also taking into account other development efforts.

Regarding your local development environment:

1. We use [golint-ci](https://golangci-lint.run/) to check the code. Consider [integrating it in your favorite IDE](https://golangci-lint.run/usage/integrations/) to avoid failing in the CI
2. **Always** run `make test` before sending a PR to make sure the license headers and the manifests are updated (and of course the unit tests are passing)
3. Consider adding a new [end-to-end](https://sdk.operatorframework.io/docs/golang/e2e-tests/) test case covering your scenario and make sure to run `make test-e2e` before sending the PR
4. Make sure to always keep your version of `Go` and the `operator-sdk` on par with the project. The current version information can be found at [the go.mod file](../go.mod)
