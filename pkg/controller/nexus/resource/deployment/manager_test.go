//     Copyright 2020 Nexus Operator and/or its authors
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

package deployment

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestNewManager_setDefaults(t *testing.T) {
	allDefaultsCommunityNexus := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{Name: "nexus-test"},
		Spec: v1alpha1.NexusSpec{
			ServiceAccountName: "nexus-test",
			Resources:          defaultResources,
			Image:              nexusCommunityLatestImage,
			LivenessProbe:      defaultProbe.DeepCopy(),
			ReadinessProbe:     defaultProbe.DeepCopy(),
		},
	}
	allDefaultsRedHatNexus := allDefaultsCommunityNexus.DeepCopy()
	allDefaultsRedHatNexus.Spec.UseRedHatImage = true
	allDefaultsRedHatNexus.Spec.Image = nexusCertifiedLatestImage

	minimumDefaultProbe := &v1alpha1.NexusProbe{
		InitialDelaySeconds: 0,
		TimeoutSeconds:      1,
		PeriodSeconds:       1,
		SuccessThreshold:    1,
		FailureThreshold:    1,
	}

	tests := []struct {
		name  string
		input *v1alpha1.Nexus
		want  *v1alpha1.Nexus
	}{
		{
			"'spec.serviceAccountName' left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ServiceAccountName = ""
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.resources' left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Resources = corev1.ResourceRequirements{}
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.useRedHatImage' set to true and 'spec.image' not left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.UseRedHatImage = true
				return nexus
			}(),
			allDefaultsRedHatNexus,
		}, {
			"'spec.useRedHatImage' set to false and 'spec.image' left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Image = ""
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.successThreshold' not equal to 1",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe.SuccessThreshold = 2
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.*' and 'spec.readinessProbe.*' don't meet minimum values",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = &v1alpha1.NexusProbe{
					InitialDelaySeconds: -1,
					TimeoutSeconds:      -1,
					PeriodSeconds:       -1,
					SuccessThreshold:    -1,
					FailureThreshold:    -1,
				}
				nexus.Spec.ReadinessProbe = nexus.Spec.LivenessProbe.DeepCopy()
				return nexus
			}(),
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = minimumDefaultProbe.DeepCopy()
				nexus.Spec.ReadinessProbe = minimumDefaultProbe.DeepCopy()
				return nexus
			}(),
		},
		{
			"Invalid 'spec.imagePullPolicy'",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ImagePullPolicy = "invalid"
				return nexus
			}(),
			allDefaultsCommunityNexus.DeepCopy(),
		},
	}

	for _, tt := range tests {
		manager := &manager{
			nexus: tt.input,
		}
		manager.setDefaults()
		if !reflect.DeepEqual(manager.nexus, tt.want) {
			t.Errorf("TestManager_setDefaults() - %s\nWant: %v\tGot: %v", tt.name, tt.want, *manager.nexus)
		}
	}
}
