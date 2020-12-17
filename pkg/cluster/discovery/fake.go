package discovery

import (
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	routev1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	discfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/rest"
	clienttesting "k8s.io/client-go/testing"

	"github.com/m88i/nexus-operator/pkg/framework/kind"
)

const openshiftGroupVersion = "openshift.io/v1"

// need to mirror this here to avoid import cycles
var m88iGroupVersion = schema.GroupVersion{Group: "apps.m88i.io", Version: "v1alpha1"}

type FakeDiscBuilder struct {
	resources []*metav1.APIResourceList
}

// NewFakeDiscBuilder returns a builder for generating a new FakeDisc
func NewFakeDiscBuilder() *FakeDiscBuilder {
	return &FakeDiscBuilder{resources: []*metav1.APIResourceList{{GroupVersion: m88iGroupVersion.String(), APIResources: []metav1.APIResource{{Kind: kind.NexusKind}}}}}
}

// OnOpenshift makes the fake client aware of resources from Openshift
func (b *FakeDiscBuilder) OnOpenshift() *FakeDiscBuilder {
	b.resources = append(b.resources,
		&metav1.APIResourceList{GroupVersion: openshiftGroupVersion},
		&metav1.APIResourceList{GroupVersion: routev1.GroupVersion.String(), APIResources: []metav1.APIResource{{Kind: kind.RouteKind}}})
	return b
}

// WithIngress makes the fake disc aware of v1 Ingresses
func (b *FakeDiscBuilder) WithIngress() *FakeDiscBuilder {
	b.resources = append(b.resources, &metav1.APIResourceList{GroupVersion: networkingv1.SchemeGroupVersion.String(), APIResources: []metav1.APIResource{{Kind: kind.IngressKind}}})
	return b
}

// WithLegacyIngress makes the fake disc aware of v1beta1 Ingresses
func (b *FakeDiscBuilder) WithLegacyIngress() *FakeDiscBuilder {
	b.resources = append(b.resources, &metav1.APIResourceList{GroupVersion: networkingv1beta1.SchemeGroupVersion.String(), APIResources: []metav1.APIResource{{Kind: kind.IngressKind}}})
	return b
}

func (b *FakeDiscBuilder) Build() *FakeDisc {
	return &FakeDisc{
		disc: &discfake.FakeDiscovery{
			Fake: &clienttesting.Fake{
				Resources: b.resources,
			},
		},
	}
}

// FakeDisc wraps a fake discovery to permit mocking errors
type FakeDisc struct {
	disc             *discfake.FakeDiscovery
	mockErr          error
	shouldClearError bool
}

// SetMockError sets the error which should be returned for the following requests
// This error will continue to be served until cleared with d.ClearMockError()
func (d *FakeDisc) SetMockError(err error) {
	d.shouldClearError = false
	d.mockErr = err
}

// SetMockErrorForOneRequest sets the error which should be returned for the following request
// this error will be set to nil after the next request
func (d *FakeDisc) SetMockErrorForOneRequest(err error) {
	d.shouldClearError = true
	d.mockErr = err
}

// ClearMockError unsets any mock errors previously set
func (d *FakeDisc) ClearMockError() {
	d.shouldClearError = false
	d.mockErr = nil
}

func (d FakeDisc) RESTClient() rest.Interface {
	return d.disc.RESTClient()
}

func (d *FakeDisc) ServerGroups() (*metav1.APIGroupList, error) {
	if d.mockErr != nil {
		if d.shouldClearError {
			defer d.ClearMockError()
		}
		return nil, d.mockErr
	}
	return d.disc.ServerGroups()
}

func (d *FakeDisc) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	if d.mockErr != nil {
		if d.shouldClearError {
			defer d.ClearMockError()
		}
		return nil, d.mockErr
	}
	return d.disc.ServerResourcesForGroupVersion(groupVersion)
}

// Deprecated: use ServerGroupsAndResources instead.
func (d *FakeDisc) ServerResources() ([]*metav1.APIResourceList, error) {
	if d.mockErr != nil {
		if d.shouldClearError {
			defer d.ClearMockError()
		}
		return nil, d.mockErr
	}
	return d.disc.ServerResources()
}

func (d *FakeDisc) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	if d.mockErr != nil {
		if d.shouldClearError {
			defer d.ClearMockError()
		}
		return nil, nil, d.mockErr
	}
	return d.disc.ServerGroupsAndResources()
}

func (d *FakeDisc) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	if d.mockErr != nil {
		if d.shouldClearError {
			defer d.ClearMockError()
		}
		return nil, d.mockErr
	}
	return d.disc.ServerPreferredResources()
}

func (d *FakeDisc) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	if d.mockErr != nil {
		if d.shouldClearError {
			defer d.ClearMockError()
		}
		return nil, d.mockErr
	}
	return d.disc.ServerPreferredNamespacedResources()
}

func (d *FakeDisc) ServerVersion() (*version.Info, error) {
	if d.mockErr != nil {
		if d.shouldClearError {
			defer d.ClearMockError()
		}
		return nil, d.mockErr
	}
	return d.disc.ServerVersion()
}

func (d *FakeDisc) OpenAPISchema() (*openapi_v2.Document, error) {
	if d.mockErr != nil {
		if d.shouldClearError {
			defer d.ClearMockError()
		}
		return nil, d.mockErr
	}
	return d.disc.OpenAPISchema()
}
