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

package test

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	discfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clienttesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	openshiftGroupVersion  = "openshift.io/v1"
	monitoringGroupVersion = "monitoring.coreos.com/v1alpha1"
)

type fakeDiscBuilder struct {
	*discfake.FakeDiscovery
}

// NewFakeClient will create a new fake client with all needed schemas
func NewFakeClient(initObjs ...runtime.Object) client.Client {
	return fake.NewFakeClientWithScheme(GetSchema(), initObjs...)
}

// GetSchema gets the needed schema for fake tests
func GetSchema() *runtime.Scheme {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.Nexus{})
	s.AddKnownTypes(routev1.GroupVersion, &routev1.Route{}, &routev1.RouteList{})
	return s
}

// NewFakeDiscoveryClient creates a fake discovery client
func NewFakeDiscoveryClient() *fakeDiscBuilder {
	return &fakeDiscBuilder{&discfake.FakeDiscovery{
		Fake: &clienttesting.Fake{
			Resources: []*metav1.APIResourceList{
				{GroupVersion: monitoringGroupVersion},
			},
		},
	}}
}

// OnOpenshift sets Openshift related server groups to the fake discovery
func (d *fakeDiscBuilder) OnOpenshift() *fakeDiscBuilder {
	d.Fake.Resources = append(d.Fake.Resources,
		&metav1.APIResourceList{GroupVersion: openshiftGroupVersion},
		&metav1.APIResourceList{GroupVersion: routev1.GroupVersion.String()})
	return d
}

// WithIngress sets Ingress the server group to the fake discovery
func (d *fakeDiscBuilder) WithIngress() *fakeDiscBuilder {
	d.Fake.Resources = append(d.Fake.Resources, &metav1.APIResourceList{GroupVersion: v1beta1.SchemeGroupVersion.String()})
	return d
}

// Build returns the fake discovery client
func (d *fakeDiscBuilder) Build() discovery.DiscoveryInterface {
	return d
}
