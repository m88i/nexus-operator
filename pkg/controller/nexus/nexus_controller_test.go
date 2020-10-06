// Copyright 2020 Nexus Operator and/or its authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nexus

import (
	"context"
	"fmt"
	"testing"

	"reflect"

	resUtils "github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	nexusres "github.com/m88i/nexus-operator/pkg/controller/nexus/resource"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/validation"
	"github.com/m88i/nexus-operator/pkg/test"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileNexus_Reconcile_NoInstance(t *testing.T) {

	// create objects to run reconcile
	cl := test.NewFakeClientBuilder().OnOpenshift().Build()
	r := newFakeReconcileNexus(cl)
	req := reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: t.Name(),
		Name:      "nexus3",
	}}

	// reconcile phase
	res, err := r.Reconcile(req)
	assert.NoError(t, err)
	assert.False(t, res.Requeue)
}

func TestReconcileNexus_Reconcile_NotPersistent(t *testing.T) {
	ns := t.Name()
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: appName},
		Spec: v1alpha1.NexusSpec{
			Replicas: 1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: false,
			},
			Networking: v1alpha1.NexusNetworking{
				Expose: true,
			},
			ServerOperations: v1alpha1.ServerOperationsOpts{DisableOperatorUserCreation: true, DisableRepositoryCreation: true},
		},
	}

	// create objects to run reconcile
	cl := test.NewFakeClientBuilder(nexus).OnOpenshift().Build()
	r := newFakeReconcileNexus(cl)
	req := reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: ns,
		Name:      appName,
	}}

	// reconcile phase
	res, err := r.Reconcile(req)
	assert.NoError(t, err)
	assert.False(t, res.Requeue)
	// let's check our replica
	dep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dep)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), *dep.Spec.Replicas)
	// shouldn't exist a pvc
	pvc := &corev1.PersistentVolumeClaim{}
	err = r.client.Get(context.TODO(), req.NamespacedName, pvc)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
	// we have routes \o/
	route := &routev1.Route{}
	err = r.client.Get(context.TODO(), req.NamespacedName, route)
	assert.NoError(t, err)
	assert.Nil(t, route.Spec.TLS)
	assert.Equal(t, route.Spec.Port.TargetPort.IntVal, dep.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)

	err = r.client.Get(context.TODO(), req.NamespacedName, nexus)
	assert.NoError(t, err)
	assert.NotNil(t, nexus)
	assert.False(t, nexus.Status.ServerOperationsStatus.ServerReady)
	assert.NotEmpty(t, nexus.Status.ServerOperationsStatus.Reason)

	// a second attempt must not requeue and not fail
	res, err = r.Reconcile(req)
	assert.NoError(t, err)
	assert.False(t, res.Requeue)
}

func TestReconcileNexus_Reconcile_Persistent(t *testing.T) {
	ns := t.Name()
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: appName},
		Spec: v1alpha1.NexusSpec{
			Replicas: 1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: true,
			},
			ServerOperations: v1alpha1.ServerOperationsOpts{DisableOperatorUserCreation: true, DisableRepositoryCreation: true},
			Networking:       v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "http://example.com"},
		},
	}

	// create objects to run reconcile
	cl := test.NewFakeClientBuilder(nexus).WithIngress().Build()
	r := newFakeReconcileNexus(cl)

	req := reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: ns,
		Name:      appName,
	}}

	// reconcile phase
	res, err := r.Reconcile(req)
	assert.NoError(t, err)
	assert.False(t, res.Requeue)
	// should exist a pvc
	pvc := &corev1.PersistentVolumeClaim{}
	err = r.client.Get(context.TODO(), req.NamespacedName, pvc)
	assert.NoError(t, err)
	assert.Equal(t, resource.MustParse("10Gi"), pvc.Spec.Resources.Requests[corev1.ResourceStorage])
	ingress := &v1beta1.Ingress{}
	err = r.client.Get(context.TODO(), req.NamespacedName, ingress)
	assert.NoError(t, err)
	assert.False(t, errors.IsNotFound(err))
}

func Test_add(t *testing.T) {
	cli := test.NewFakeClientBuilder().Build()
	mgr := test.NewManager(cli)
	r := newFakeReconcileNexus(cli)
	err := add(mgr, &r)
	assert.NoError(t, err)
}

func TestReconcileNexus_handleUpdate(t *testing.T) {
	r := newFakeReconcileNexus(test.NewFakeClientBuilder().Build())
	baseNexus := validation.AllDefaultsCommunityNexus.DeepCopy()
	baseNexus.Spec.AutomaticUpdate.Disabled = true
	deploymentType := reflect.TypeOf(appsv1.Deployment{})

	// First, let's test the first deployment scenario, which isn't an update
	requiredRes := make(map[reflect.Type][]resUtils.KubernetesResource)
	deployedRes := make(map[reflect.Type][]resUtils.KubernetesResource)

	requiredDep := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: fmt.Sprintf("%s:%s", baseNexus.Spec.Image, "3.25.1"),
						},
					},
				},
			},
		},
	}
	requiredRes[deploymentType] = []resUtils.KubernetesResource{requiredDep}
	assert.NoError(t, r.handleUpdate(baseNexus, requiredRes, deployedRes))

	// Now let's test with an existing deployment
	deployedDep := requiredDep.DeepCopy()
	deployedRes[deploymentType] = []resUtils.KubernetesResource{deployedDep}
	assert.NoError(t, r.handleUpdate(baseNexus, requiredRes, deployedRes))
}

func newFakeReconcileNexus(cl *test.FakeClient) ReconcileNexus {
	return ReconcileNexus{
		client:             cl,
		scheme:             scheme.Scheme,
		discoveryClient:    cl,
		resourceSupervisor: nexusres.NewSupervisor(cl, cl),
	}
}
