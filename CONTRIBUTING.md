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
4. Make sure to always keep your version of `Go` and the `operator-sdk` on par with the project. The current version information can be found at [the go.mod file](go.mod)

## E2E Testing

If you added a new functionality and are willing to add some end-to-end (E2E) testing of your own, please add a test case to `test/e2e/nexus_test.go`.

The test case structure allows you to name your test appropriately (try naming it in a way it's clear what it's testing), provide a Nexus CR that the Operator will use to generate the other resources, provide additional checks your feature may require and provide a custom cleanup function if necessary.

Then each test case is submitted to a series of checks which should make sure everything on the cluster is as it should, based on the Nexus CR that has been defined.

Let's take our smoke test as an example to go over the test cases structure:

```go
testCases := []struct {
	name             string (1)
	input            *v1alpha1.Nexus (2)
	cleanup          func() error (3)
	additionalChecks []func(nexus *v1alpha1.Nexus) error (4)
}{
	{
		name: "Smoke test: no persistence, nodeport exposure", (1)
		input: &v1alpha1.Nexus{ (2)
			ObjectMeta: metav1.ObjectMeta{
				Name:      nexusName,
				Namespace: namespace,
			},
			Spec: defaultNexusSpec, (5)
		},
		cleanup: tester.defaultCleanup, (3)
		additionalChecks: nil, (4)
	},
```

> (1): the test case's name. In this scenario we're testing a deployment with all default values, no persistence and exposed via Node Port.<br>
> (2): the Nexus CR which the Operator will use to orchestrate and maintain your Nexus3 deployment<br>
> (3): a cleanup function which should be ran after the test has been completed<br>
> (4): additional checks your test case may need<br>
> (5): the base, default Nexus CR specification which should be used for testing. Modify this to test your own features<br>

**Important**: although the operator will set the defaults on the Nexus CR you provide it with, the tests will use your original CR for comparison, so be sure to make a completely valid Nexus CR for your test case as it *will not* be modified to insert default values.

### Custom Nexus CR

If your test requires modifications to the default Nexus CR, you can do so directly and concisely when defining the test case by making use of anonymous functions.

For example:

```go
{
    name: "Networking: ingress with no TLS",
    input: &v1alpha1.Nexus{
        ObjectMeta: metav1.ObjectMeta{
            Name:      nexusName,
            Namespace: namespace,
        },
        Spec: func() v1alpha1.NexusSpec {
            spec := *defaultNexusSpec.DeepCopy()
            spec.Networking = v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "test-example.com"}
            return spec
        }(),
    },
    cleanup: tester.defaultCleanup,
    additionalChecks: nil,
},
```

When defining the Nexus's specification in this case we're actually calling an anonymous function that acquires the default spec, modifies the required fields and then returns that spec, thus making the necessary changes for the test.

### Custom Cleanup functions

Our test cases make use of [functions first-class citizenship in Go](https://golang.org/doc/codewalk/functions/) by declaring the cleanup function as a field from the test case. This way it's possible to specify our own custom cleanup function for a test.

In previous examples, `tester.defaultCleanup` was used, which simply deletes all Nexus CRs in the namespace, but you may want to do some additional computation when cleaning up, such as counting to 5 (intentionally useless to promote simplicity in this example):

```go
{
    name: "Test Example: this counts to 5 during cleanup and uses the default cleanup once done",
    input: &v1alpha1.Nexus{
        ObjectMeta: metav1.ObjectMeta{
            Name:      nexusName,
            Namespace: namespace,
        },
        Spec: defaultNexusSpec,
    },
    cleanup: func() error {
        for i := 0; i < 5; i++ {
            tester.t.Logf("Count: %d", i)
        }
        return tester.defaultCleanup()
    },
    additionalChecks: nil,
},
```

It's possible, of course, to not use the default cleanup function at all, but be sure to actually delete the resources you created if they conflict with other test cases (the framework itself will delete the whole namespace once the tests are done):

```go
{
	name: "Test Example: this only counts to 5 during cleanup and does not delete anything",
	input: &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nexusName,
			Namespace: namespace,
		},
		Spec: defaultNexusSpec,
	},
	cleanup: func() error {
		for i := 0; i < 5; i++ {
			tester.t.Logf("Count: %d", i)
		}
		return nil
	},
	additionalChecks: nil,
},
```

### Running additional checks

If your testing needs to check something that isn't already checked by default you may add functions to perform these checks as the function that is responsible for running the default checks will receive them as [variadic arguments](https://gobyexample.com/variadic-functions).

In another useless yet simple example, let's also make sure that 5 is greater than 4 when performing our checks:

```go
{
	name: "Test Example: this will also check if 5 > 4",
	input: &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nexusName,
			Namespace: namespace,
		},
		Spec: defaultNexusSpec,
	},
	cleanup: tester.defaultCleanup,
	additionalChecks: []func(nexus *v1alpha1.Nexus)error{
		func(nexus *v1alpha1.Nexus) error {
			assert.Greater(tester.t, 5, 4)
			return nil
		},
	},
},
```