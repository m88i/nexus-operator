package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/m88i/nexus-operator/pkg/apis"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/stretchr/testify/assert"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 10
	timeout              = time.Second * 360
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
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

func nexusNoPersitenceNodePort(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	// create nexus custom resource
	nexus3 := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nexus3",
			Namespace: namespace,
		},
		Spec: v1alpha1.NexusSpec{
			Replicas:       1,
			Networking:     v1alpha1.NexusNetworking{Expose: true, NodePort: 31031, ExposeAs: v1alpha1.NodePortExposeType},
			Persistence:    v1alpha1.NexusPersistence{Persistent: false},
			UseRedHatImage: false,
			LivenessProbe:  &v1alpha1.NexusProbe{InitialDelaySeconds: 240},
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), nexus3, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	// wait for nexus3 to finish deployment
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "nexus3", 1, retryInterval, timeout)
	assert.NoError(t, err)

	// check the service
	svc := &v1.Service{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: nexus3.Name, Namespace: nexus3.Namespace}, svc)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, nexus3.Spec.Networking.NodePort, svc.Spec.Ports[0].NodePort)

	// we don't have a PVC created
	pvc := &v1.PersistentVolumeClaim{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: nexus3.Name, Namespace: nexus3.Namespace}, pvc)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))

	return nil
}

func NexusCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
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

	if err = nexusNoPersitenceNodePort(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}
