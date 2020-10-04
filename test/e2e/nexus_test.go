package e2e

import (
	"testing"
	"time"

	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/validation"
	corev1 "k8s.io/api/core/v1"

	"github.com/m88i/nexus-operator/pkg/apis"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const nexusName = "nexus3"

var (
	retryInterval        = time.Second * 30
	timeout              = time.Second * 720
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5

	defaultNexusSpec = func() v1alpha1.NexusSpec {
		spec := v1alpha1.NexusSpec{
			AutomaticUpdate:             v1alpha1.NexusAutomaticUpdate{Disabled: true},
			Replicas:                    1,
			Image:                       validation.NexusCommunityImage,
			ImagePullPolicy:             corev1.PullIfNotPresent,
			Resources:                   validation.DefaultResources,
			Persistence:                 v1alpha1.NexusPersistence{Persistent: false},
			UseRedHatImage:              false,
			GenerateRandomAdminPassword: false,
			Networking:                  v1alpha1.NexusNetworking{Expose: true, NodePort: 31031, ExposeAs: v1alpha1.NodePortExposeType},
			ServiceAccountName:          nexusName,
			LivenessProbe:               validation.DefaultProbe.DeepCopy(),
			ReadinessProbe:              validation.DefaultProbe.DeepCopy(),
		}
		spec.LivenessProbe.InitialDelaySeconds = 480
		return spec
	}()
)

func TestNexus(t *testing.T) {
	nexusList := &v1alpha1.NexusList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, nexusList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("nexus-group", func(t *testing.T) {
		t.Run("Cluster", NexusCluster)
	})
}

func NexusCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewContext(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for nexus-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "nexus-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	tester := &tester{
		t:   t,
		f:   f,
		ctx: ctx,
	}

	testCases := []struct {
		name             string
		input            *v1alpha1.Nexus
		cleanup          func() error
		additionalChecks []func(nexus *v1alpha1.Nexus) error
	}{
		{
			name: "Smoke test: no persistence, nodeport exposure",
			input: &v1alpha1.Nexus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nexusName,
					Namespace: namespace,
				},
				Spec: defaultNexusSpec,
			},
			cleanup:          tester.defaultCleanup,
			additionalChecks: nil,
		},
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
			cleanup:          tester.defaultCleanup,
			additionalChecks: nil,
		},
		{
			name: "Networking: ingress with TLS secret",
			input: &v1alpha1.Nexus{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nexusName,
					Namespace: namespace,
				},
				Spec: func() v1alpha1.NexusSpec {
					spec := *defaultNexusSpec.DeepCopy()
					spec.Networking = v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "test-example.com", TLS: v1alpha1.NexusNetworkingTLS{SecretName: "test-secret"}}
					return spec
				}(),
			},
			cleanup:          tester.defaultCleanup,
			additionalChecks: nil,
		},
	}

	for _, testCase := range testCases {
		tester.t.Logf("Test: %s\nInput: %+v", testCase.name, testCase.input)
		if err = tester.runChecks(testCase.input, testCase.additionalChecks...); err != nil {
			tester.t.Fatal(err)
		}
		if err = testCase.cleanup(); err != nil {
			tester.t.Logf("%v", err)
		}
	}
}
