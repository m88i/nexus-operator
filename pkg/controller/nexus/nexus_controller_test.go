//     Copyright 2019 Nexus Operator and/or its authors
//
//     This file is part of Nexus Operator.
//
//     Nexus Operator is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     Nexus Operator is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with Nexus Operator.  If not, see <https://www.gnu.org/licenses/>.

package nexus

import (
	"context"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	nexusres "github.com/m88i/nexus-operator/pkg/controller/nexus/resource"
	"github.com/m88i/nexus-operator/pkg/test"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestReconcileNexus_Reconcile_NotPersistent(t *testing.T) {
	ns := t.Name()
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: v1.ObjectMeta{Namespace: ns, Name: appName},
		Spec: v1alpha1.NexusSpec{
			Replicas: 1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: false,
			},
			Networking: v1alpha1.NexusNetworking{
				Expose: true,
			},
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
}

func TestReconcileNexus_Reconcile_Persistent(t *testing.T) {
	ns := t.Name()
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: v1.ObjectMeta{Namespace: ns, Name: appName},
		Spec: v1alpha1.NexusSpec{
			Replicas: 1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: true,
			},
		},
	}

	// create objects to run reconcile
	cl := test.NewFakeClientBuilder(nexus).Build()
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
	// networking is disabled
	ingress := &v1beta1.Ingress{}
	err = r.client.Get(context.TODO(), req.NamespacedName, ingress)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
}

func newFakeReconcileNexus(cl *test.FakeClient) ReconcileNexus {
	return ReconcileNexus{
		client:             cl,
		scheme:             scheme.Scheme,
		discoveryClient:    cl,
		resourceSupervisor: nexusres.NewSupervisor(cl, cl),
	}
}
