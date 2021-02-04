# Contributing to Nexus Operator

## Have found a bug or have a feature request?

Please, open an issue for us. That will really help us improving the Operator and it would benefit other users such as
yourself. There are templates ready for bug reporting and feature request to make things easier for you.

## Have any questions?

We're happy to answer! Either [open an issue](https://github.com/m88i/nexus-operator/issues) or send an email to our
mailing list: [nexus-operator@googlegroups.com](mailto:nexus-operator@googlegroups.com).

## Are you willing to send a PR?

Before sending a PR, consider opening an issue first. This way we can discuss your approach and the motivations behind
it while also taking into account other development efforts.

Regarding your local development environment:

1. We use [golint-ci](https://golangci-lint.run/) to check the code.
   Consider [integrating it in your favorite IDE](https://golangci-lint.run/usage/integrations/) to avoid failing in the
   CI
2. **Always** run `make test` before sending a PR to make sure the license headers and the manifests are updated (and of
   course the unit tests are passing)
3. Consider adding a new [end-to-end](https://sdk.operatorframework.io/docs/building-operators/golang/testing/) test
   case covering your scenario in `controllers/nexus_controller_test.go` and make sure to run `make test` before sending
   the PR
4. Make sure to always keep your version of `go` and the `operator-sdk` on par with the project.

    - go 1.15
    - operator-sdk v1.2.0

To run all tests and push, all in one go, you may run `make pr-prep`. If any tests fail or if the build fails, the
process will be terminated so that you make the necessary adjustments. If they are all successful, you'll be prompted to
push your committed changes.

```shell
$ make pr-prep
# (output omitted)
All tests were successful!
Do you wish to push? (y/n) y
Insert the remote name: [origin] 
Insert branch: [pr-prep] 
Pushing to origin/pr-prep
# (output omitted)
```

If you don't inform remote name and branch, it will use "origin" as the remote and your current branch (the defaults,
which appear between "[]"). Double check if the information is correct.

If you don't want to go over the interactive prompt every time, you can push with the defaults using the
`PUSH_WITH_DEFAULTS` environment variable:

```shell
$ PUSH_WITH_DEFAULTS=TRUE make pr-prep
# (output omitted)
All tests were successful!
Pushing to origin/pr-prep
# (output omitted)
```

## E2E Testing

If you added a new functionality and are willing to add some end-to-end (E2E) testing of your own, please add a test
case to `controllers/nexus_controller_test.go`.

All mutating (including default values) and validating logic lives in our admission webhooks, which are not in place
when the e2e suite runs. Because of that, no changes or checks are performed against Nexus CRs being reconciled. Be sure
to always write tests which use valid Nexus resources.

In the future, after we migrate to operator-sdk v1.3.0 and kubebuilder plugin v3, scaffolding will create a separate
suite for testing the webhooks.