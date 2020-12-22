package v1alpha1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/pkg/cluster/discovery"
	"github.com/m88i/nexus-operator/pkg/framework"
)

func TestNewMutator(t *testing.T) {
	tests := []struct {
		name   string
		client *discovery.FakeDisc
		want   *validator
	}{
		{
			"On OCP",
			discovery.NewFakeDiscBuilder().OnOpenshift().Build(),
			&validator{
				routeAvailable:   true,
				ingressAvailable: false,
			},
		},
		{
			"On K8s, no ingress",
			discovery.NewFakeDiscBuilder().Build(),
			&validator{
				routeAvailable:   false,
				ingressAvailable: false,
			},
		},
		{
			"On K8s with v1beta1 ingress",
			discovery.NewFakeDiscBuilder().WithLegacyIngress().Build(),
			&validator{
				routeAvailable:   false,
				ingressAvailable: true,
			},
		},
		{
			"On K8s with v1 ingress",
			discovery.NewFakeDiscBuilder().WithIngress().Build(),
			&validator{
				routeAvailable:   false,
				ingressAvailable: true,
			},
		},
	}

	// the nexus itself isn't important, but we need it to avoid nil dereference
	n := &Nexus{ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: t.Name()}}
	for _, tt := range tests {
		discovery.SetClient(tt.client)
		m := newMutator(n)
		assert.Equal(t, tt.want.ingressAvailable, m.ingressAvailable)
		assert.Equal(t, tt.want.routeAvailable, m.routeAvailable)
	}

	// now let's test out error
	client := discovery.NewFakeDiscBuilder().WithIngress().Build()
	errString := "test error"
	client.SetMockError(fmt.Errorf(errString))
	discovery.SetClient(client)
	got := newMutator(n)
	// all calls to discovery failed, so all should be false
	assert.False(t, got.routeAvailable)
	assert.False(t, got.ingressAvailable)
}

func TestMutator_mutate_AutomaticUpdate(t *testing.T) {
	discovery.SetClient(discovery.NewFakeDiscBuilder().Build())
	nexus := &Nexus{Spec: NexusSpec{AutomaticUpdate: NexusAutomaticUpdate{}}}
	nexus.Spec.Image = NexusCommunityImage
	newMutator(nexus).mutate()
	latestMinor, err := framework.GetLatestMinor()
	if err != nil {
		// If we couldn't fetch the tags updates should be disabled
		assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)
	} else {
		assert.Equal(t, latestMinor, *nexus.Spec.AutomaticUpdate.MinorVersion)
	}

	// Now an invalid image
	nexus = &Nexus{Spec: NexusSpec{AutomaticUpdate: NexusAutomaticUpdate{}}}
	nexus.Spec.Image = "some-image"
	newMutator(nexus).mutate()
	assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)

	// Informed a minor which does not exist
	nexus = &Nexus{Spec: NexusSpec{AutomaticUpdate: NexusAutomaticUpdate{}}}
	nexus.Spec.Image = NexusCommunityImage
	bogusMinor := -1
	nexus.Spec.AutomaticUpdate.MinorVersion = &bogusMinor
	newMutator(nexus).mutate()
	latestMinor, err = framework.GetLatestMinor()
	if err != nil {
		// If we couldn't fetch the tags updates should be disabled
		assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)
	} else {
		assert.Equal(t, latestMinor, *nexus.Spec.AutomaticUpdate.MinorVersion)
	}
}

func TestMutator_mutate_Networking(t *testing.T) {
	tests := []struct {
		name  string
		disc  *discovery.FakeDisc
		input *Nexus
		want  *Nexus
	}{
		{
			"'spec.networking.exposeAs' left blank on OCP",
			discovery.NewFakeDiscBuilder().OnOpenshift().Build(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				n.Spec.Networking.ExposeAs = RouteExposeType
				return n
			}(),
		},
		{
			"'spec.networking.exposeAs' left blank on K8s with ingress",
			discovery.NewFakeDiscBuilder().WithIngress().Build(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				n.Spec.Networking.ExposeAs = IngressExposeType
				return n
			}(),
		},
		{
			"'spec.networking.exposeAs' left blank on K8s, but Ingress unavailable",
			discovery.NewFakeDiscBuilder().Build(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				n.Spec.Networking.ExposeAs = NodePortExposeType
				return n
			}(),
		},
	}

	for _, tt := range tests {
		discovery.SetClient(tt.disc)
		newMutator(tt.input).mutate()
		assert.Equal(t, tt.want, tt.input)
	}
}

func TestMutator_mutate_Persistence(t *testing.T) {
	tests := []struct {
		name  string
		input *Nexus
		want  *Nexus
	}{
		{
			"'spec.persistence.volumeSize' left blank",
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Persistence.Persistent = true
				return n
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Persistence.Persistent = true
				n.Spec.Persistence.VolumeSize = DefaultVolumeSize
				return n
			}(),
		},
	}

	discovery.SetClient(discovery.NewFakeDiscBuilder().Build())
	for _, tt := range tests {
		newMutator(tt.input).mutate()
		assert.Equal(t, tt.want, tt.input)
	}
}

func TestMutator_mutate_Security(t *testing.T) {
	tests := []struct {
		name  string
		input *Nexus
		want  *Nexus
	}{
		{
			"'spec.serviceAccountName' left blank",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ServiceAccountName = ""
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
	}

	discovery.SetClient(discovery.NewFakeDiscBuilder().Build())
	for _, tt := range tests {
		newMutator(tt.input).mutate()
		assert.Equal(t, tt.want, tt.input)
	}
}

func TestMutator_mutate_Deployment(t *testing.T) {
	minimumDefaultProbe := &NexusProbe{
		InitialDelaySeconds: 0,
		TimeoutSeconds:      1,
		PeriodSeconds:       1,
		SuccessThreshold:    1,
		FailureThreshold:    1,
	}

	tests := []struct {
		name  string
		input *Nexus
		want  *Nexus
	}{
		{
			"'spec.resources' left blank",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Resources = v1.ResourceRequirements{}
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
		{
			"'spec.useRedHatImage' set to true and 'spec.image' not left blank",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.UseRedHatImage = true
				nexus.Spec.Image = "some-image"
				return nexus
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.UseRedHatImage = true
				n.Spec.Image = NexusCertifiedImage
				return n
			}(),
		},
		{
			"'spec.useRedHatImage' set to false and 'spec.image' left blank",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Image = ""
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.successThreshold' not equal to 1",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe.SuccessThreshold = 2
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.*' and 'spec.readinessProbe.*' don't meet minimum values",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = &NexusProbe{
					InitialDelaySeconds: -1,
					TimeoutSeconds:      -1,
					PeriodSeconds:       -1,
					SuccessThreshold:    -1,
					FailureThreshold:    -1,
				}
				nexus.Spec.ReadinessProbe = nexus.Spec.LivenessProbe.DeepCopy()
				return nexus
			}(),
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = minimumDefaultProbe.DeepCopy()
				nexus.Spec.ReadinessProbe = minimumDefaultProbe.DeepCopy()
				return nexus
			}(),
		},
		{
			"Unset 'spec.livenessProbe' and 'spec.readinessProbe'",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = nil
				nexus.Spec.ReadinessProbe = nil
				return nexus
			}(),
			AllDefaultsCommunityNexus.DeepCopy(),
		},
		{
			"Invalid 'spec.imagePullPolicy'",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ImagePullPolicy = "invalid"
				return nexus
			}(),
			AllDefaultsCommunityNexus.DeepCopy(),
		},
		{
			"Invalid 'spec.replicas'",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Replicas = 3
				return nexus
			}(),
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Replicas = 1
				return nexus
			}(),
		},
	}

	discovery.SetClient(discovery.NewFakeDiscBuilder().Build())
	for _, tt := range tests {
		newMutator(tt.input).mutate()
		assert.Equal(t, tt.want, tt.input)
	}
}
