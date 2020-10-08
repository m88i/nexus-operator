package e2e

import (
	"context"
	"reflect"
	"testing"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/deployment"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/server"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type tester struct {
	t   *testing.T
	f   *test.Framework
	ctx *test.Context
}

func (tester *tester) defaultCleanup() error {
	namespace, err := tester.ctx.GetOperatorNamespace()
	if err != nil {
		return err
	}
	return tester.f.Client.DeleteAllOf(context.TODO(), &v1alpha1.Nexus{}, client.InNamespace(namespace))
}

func (tester *tester) runChecks(nexus *v1alpha1.Nexus, additionalChecks ...func(nexus *v1alpha1.Nexus) error) error {
	err := tester.createAndWait(nexus)
	if err != nil {
		return err
	}

	tester.checkService(nexus)
	tester.checkPVC(nexus)
	tester.checkDeployment(nexus)
	tester.checkServiceAccount(nexus)
	tester.checkIngress(nexus)
	tester.checkServerInteraction(nexus)

	for _, check := range additionalChecks {
		if err = check(nexus); err != nil {
			return err
		}
	}
	return nil
}

func (tester *tester) createAndWait(nexus *v1alpha1.Nexus) error {
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err := tester.f.Client.Create(context.TODO(), nexus, &test.CleanupOptions{TestContext: tester.ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	// wait for nexus3 to finish deployment
	err = e2eutil.WaitForDeployment(tester.t, tester.f.KubeClient, nexus.Namespace, nexus.Name, int(nexus.Spec.Replicas), retryInterval, timeout)
	assert.NoError(tester.t, err)
	return nil
}

func (tester *tester) checkService(nexus *v1alpha1.Nexus) {
	svc := &corev1.Service{}
	err := tester.f.Client.Get(context.TODO(), types.NamespacedName{Name: nexus.Name, Namespace: nexus.Namespace}, svc)
	assert.NoError(tester.t, err)
	assert.NotNil(tester.t, svc)

	assert.Equal(tester.t, deployment.DefaultHTTPPort, int(svc.Spec.Ports[0].Port))
	assert.Equal(tester.t, deployment.NexusServicePort, int(svc.Spec.Ports[0].TargetPort.IntVal))
	assert.Equal(tester.t, deployment.NexusPortName, svc.Spec.Ports[0].Name)

	if nexus.Spec.Networking.ExposeAs == v1alpha1.NodePortExposeType {
		assert.Equal(tester.t, nexus.Spec.Networking.NodePort, svc.Spec.Ports[0].NodePort)
		assert.Equal(tester.t, corev1.ServiceTypeNodePort, svc.Spec.Type)
		assert.Equal(tester.t, corev1.ServiceExternalTrafficPolicyTypeCluster, svc.Spec.ExternalTrafficPolicy)
	}
}

func (tester *tester) checkPVC(nexus *v1alpha1.Nexus) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := tester.f.Client.Get(context.TODO(), types.NamespacedName{Name: nexus.Name, Namespace: nexus.Namespace}, pvc)

	if !nexus.Spec.Persistence.Persistent {
		assert.NotNil(tester.t, err)
		assert.True(tester.t, errors.IsNotFound(err))
		return
	}

	assert.Nil(tester.t, err)
	assert.Equal(tester.t, resource.MustParse(nexus.Spec.Persistence.VolumeSize), pvc.Spec.Resources.Requests[corev1.ResourceStorage])

	if nexus.Spec.Replicas > 1 {
		assert.Equal(tester.t, corev1.ReadWriteMany, pvc.Spec.AccessModes[0])
	} else {
		assert.Equal(tester.t, corev1.ReadWriteOnce, pvc.Spec.AccessModes[0])
	}

	if len(nexus.Spec.Persistence.StorageClass) > 0 {
		assert.Equal(tester.t, nexus.Spec.Persistence.StorageClass, *pvc.Spec.StorageClassName)
	}
}

func (tester *tester) checkDeployment(nexus *v1alpha1.Nexus) {
	d := &appsv1.Deployment{}
	err := tester.f.Client.Get(context.TODO(), types.NamespacedName{Name: nexus.Name, Namespace: nexus.Namespace}, d)
	assert.NoError(tester.t, err)

	assert.True(tester.t, reflect.DeepEqual(d.Spec.Template.Spec.Containers[0].Resources.Requests, nexus.Spec.Resources.Requests))
	assert.True(tester.t, reflect.DeepEqual(d.Spec.Template.Spec.Containers[0].Resources.Limits, nexus.Spec.Resources.Limits))

	assert.Equal(tester.t, nexus.Spec.Replicas, *d.Spec.Replicas)
	assert.Equal(tester.t, nexus.Spec.ServiceAccountName, d.Spec.Template.Spec.ServiceAccountName)
	assert.Equal(tester.t, nexus.Spec.Image, d.Spec.Template.Spec.Containers[0].Image)
	assert.True(tester.t, tester.containsJvmEnv(d))

	if nexus.Spec.Persistence.Persistent {
		assert.NotEmpty(tester.t, d.Spec.Template.Spec.Volumes)
		assert.NotEmpty(tester.t, d.Spec.Template.Spec.Containers[0].VolumeMounts)
	} else {
		assert.Empty(tester.t, d.Spec.Template.Spec.Volumes)
		assert.Empty(tester.t, d.Spec.Template.Spec.Containers[0].VolumeMounts)
	}

	// here we only check if the security context has been set as we don't want to couple the business logic (and the actual values) to this test
	// testing the logic is up to the unit test
	if nexus.Spec.UseRedHatImage {
		assert.Nil(tester.t, d.Spec.Template.Spec.SecurityContext)
	} else {
		assert.NotNil(tester.t, d.Spec.Template.Spec.SecurityContext)
	}

	tester.checkDeploymentProbes(nexus, d)
}

func (tester *tester) containsJvmEnv(d *appsv1.Deployment) bool {
	// here we only check if the parameters have been set as we don't want to couple the calculation logic to this test
	// testing the logic is up to the unit test
	for _, env := range d.Spec.Template.Spec.Containers[0].Env {
		if env.Name == deployment.JvmArgsEnvKey {
			return true
		}
	}
	return false
}

func (tester *tester) checkDeploymentProbes(nexus *v1alpha1.Nexus, d *appsv1.Deployment) {
	assert.Equal(tester.t, nexus.Spec.LivenessProbe.FailureThreshold, d.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold)
	assert.Equal(tester.t, nexus.Spec.LivenessProbe.PeriodSeconds, d.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds)
	assert.Equal(tester.t, nexus.Spec.LivenessProbe.SuccessThreshold, d.Spec.Template.Spec.Containers[0].LivenessProbe.SuccessThreshold)
	assert.Equal(tester.t, nexus.Spec.LivenessProbe.TimeoutSeconds, d.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds)
	assert.Equal(tester.t, nexus.Spec.LivenessProbe.InitialDelaySeconds, d.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds)

	assert.Equal(tester.t, nexus.Spec.ReadinessProbe.FailureThreshold, d.Spec.Template.Spec.Containers[0].ReadinessProbe.FailureThreshold)
	assert.Equal(tester.t, nexus.Spec.ReadinessProbe.PeriodSeconds, d.Spec.Template.Spec.Containers[0].ReadinessProbe.PeriodSeconds)
	assert.Equal(tester.t, nexus.Spec.ReadinessProbe.SuccessThreshold, d.Spec.Template.Spec.Containers[0].ReadinessProbe.SuccessThreshold)
	assert.Equal(tester.t, nexus.Spec.ReadinessProbe.TimeoutSeconds, d.Spec.Template.Spec.Containers[0].ReadinessProbe.TimeoutSeconds)
	assert.Equal(tester.t, nexus.Spec.ReadinessProbe.InitialDelaySeconds, d.Spec.Template.Spec.Containers[0].ReadinessProbe.InitialDelaySeconds)
}

func (tester *tester) checkServiceAccount(nexus *v1alpha1.Nexus) {
	account := &corev1.ServiceAccount{}
	err := tester.f.Client.Get(context.TODO(), types.NamespacedName{Name: nexus.Name, Namespace: nexus.Namespace}, account)
	assert.Nil(tester.t, err)
}

func (tester *tester) checkIngress(nexus *v1alpha1.Nexus) {
	if !nexus.Spec.Networking.Expose {
		return
	}

	ingress := &v1beta1.Ingress{}
	err := tester.f.Client.Get(context.TODO(), types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name}, ingress)

	if nexus.Spec.Networking.ExposeAs != v1alpha1.IngressExposeType {
		assert.NotNil(tester.t, err)
		assert.True(tester.t, errors.IsNotFound(err))
		return
	}

	assert.Nil(tester.t, err)
	assert.Equal(tester.t, nexus.Spec.Networking.Host, ingress.Spec.Rules[0].Host)
	assert.True(tester.t, tester.ingressPointsToCorrectService(nexus, ingress))

	if len(nexus.Spec.Networking.TLS.SecretName) > 0 {
		assert.Equal(tester.t, nexus.Spec.Networking.TLS.SecretName, ingress.Spec.TLS[0].SecretName)
		assert.Contains(tester.t, ingress.Spec.TLS[0].Hosts, nexus.Spec.Networking.Host)
	}
}

func (tester *tester) ingressPointsToCorrectService(nexus *v1alpha1.Nexus, ingress *v1beta1.Ingress) bool {
	for _, path := range ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths {
		if path.Backend.ServiceName == nexus.Name {
			return true
		}
	}
	return false
}

// checkServerInteraction verifies if the secret holding operator user credentials has the credentials or not
func (tester *tester) checkServerInteraction(nexus *v1alpha1.Nexus) {
	tester.t.Log("Checking Nexus Server interactions")
	if nexus.Spec.ServerOperations.DisableOperatorUserCreation || nexus.Spec.GenerateRandomAdminPassword {
		tester.t.Log("Nexus Server interactions disabled, skipping")
		return
	}
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		secret := &corev1.Secret{ObjectMeta: v1.ObjectMeta{Name: nexus.Name, Namespace: nexus.Namespace}}
		err = tester.f.Client.Get(context.TODO(), framework.Key(nexus), secret)
		if err != nil {
			if errors.IsNotFound(err) {
				tester.t.Log("Waiting for Nexus secret to be available")
				return false, nil
			}
			return false, err
		}
		if len(secret.Data[server.SecretKeyUsername]) > 0 && len(secret.Data[server.SecretKeyPassword]) > 0 {
			tester.t.Log("Nexus Operator credentials found! Test OK.")
			return true, nil
		}
		tester.t.Log("Waiting for Nexus secret to have credentials")
		return false, nil
	})
	assert.NoError(tester.t, err)
}
