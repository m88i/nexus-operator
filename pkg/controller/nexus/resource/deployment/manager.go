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
	ctx "context"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/logger"
	"github.com/m88i/nexus-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8sres "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NexusCommunityLatestImage       = "docker.io/sonatype/nexus3:latest"
	NexusCertifiedLatestImage       = "registry.connect.redhat.com/sonatype/nexus-repository-manager"
	mgrNotInit                      = "the manager has not been initialized"
	probeDefaultInitialDelaySeconds = int32(240)
	probeDefaultTimeoutSeconds      = int32(15)
	probeDefaultPeriodSeconds       = int32(10)
	probeDefaultSuccessThreshold    = int32(1)
	probeDefaultFailureThreshold    = int32(3)
)

var (
	DefaultResources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    k8sres.MustParse("2"),
			corev1.ResourceMemory: k8sres.MustParse("2Gi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    k8sres.MustParse("1"),
			corev1.ResourceMemory: k8sres.MustParse("2Gi"),
		},
	}

	DefaultProbe = &v1alpha1.NexusProbe{
		InitialDelaySeconds: probeDefaultInitialDelaySeconds,
		TimeoutSeconds:      probeDefaultTimeoutSeconds,
		PeriodSeconds:       probeDefaultPeriodSeconds,
		SuccessThreshold:    probeDefaultSuccessThreshold,
		FailureThreshold:    probeDefaultFailureThreshold,
	}

	log = logger.GetLogger("deployment_manager")
)

// Manager is responsible for creating deployment-related resources, fetching deployed ones and comparing them
type Manager struct {
	nexus  *v1alpha1.Nexus
	client client.Client
}

// NewManager creates a deployment resources manager
func NewManager(nexus v1alpha1.Nexus, client client.Client) *Manager {
	mgr := &Manager{
		nexus:  &nexus,
		client: client,
	}
	mgr.setDefaults()
	return mgr
}

// setDefaults destructively sets default for unset values or ensure valid ranges for set values in the Nexus CR
func (m *Manager) setDefaults() {
	m.setSvcAccntDefaults()
	m.setResourcesDefaults()
	m.setImageDefaults()
	m.setProbeDefaults()
}

func (m *Manager) setSvcAccntDefaults() {
	if len(m.nexus.Spec.ServiceAccountName) == 0 {
		m.nexus.Spec.ServiceAccountName = m.nexus.Name
	}
}

func (m *Manager) setResourcesDefaults() {
	if m.nexus.Spec.Resources.Requests == nil && m.nexus.Spec.Resources.Limits == nil {
		m.nexus.Spec.Resources = DefaultResources
	}
}

func (m *Manager) setImageDefaults() {
	if m.nexus.Spec.UseRedHatImage {
		if len(m.nexus.Spec.Image) > 0 {
			log.Warnf("Nexus CR configured to the use Red Hat Certified Image, ignoring 'spec.image' field.")
		}
		m.nexus.Spec.Image = NexusCertifiedLatestImage
	} else if len(m.nexus.Spec.Image) == 0 {
		m.nexus.Spec.Image = NexusCommunityLatestImage
	}

	if len(m.nexus.Spec.ImagePullPolicy) > 0 &&
		m.nexus.Spec.ImagePullPolicy != corev1.PullAlways &&
		m.nexus.Spec.ImagePullPolicy != corev1.PullIfNotPresent &&
		m.nexus.Spec.ImagePullPolicy != corev1.PullNever {

		log.Warnf("Invalid 'spec.imagePullPolicy', unsetting the value. The pull policy will be determined by the image tag. Consider setting this value to '%s', '%s' or '%s'", corev1.PullAlways, corev1.PullIfNotPresent, corev1.PullNever)
		m.nexus.Spec.ImagePullPolicy = ""
	}
}

func (m *Manager) setProbeDefaults() {
	if m.nexus.Spec.LivenessProbe != nil {
		m.nexus.Spec.LivenessProbe.FailureThreshold =
			util.EnsureMinimum(m.nexus.Spec.LivenessProbe.FailureThreshold, 1)
		m.nexus.Spec.LivenessProbe.InitialDelaySeconds =
			util.EnsureMinimum(m.nexus.Spec.LivenessProbe.InitialDelaySeconds, 0)
		m.nexus.Spec.LivenessProbe.PeriodSeconds =
			util.EnsureMinimum(m.nexus.Spec.LivenessProbe.PeriodSeconds, 1)
		m.nexus.Spec.LivenessProbe.TimeoutSeconds =
			util.EnsureMinimum(m.nexus.Spec.LivenessProbe.TimeoutSeconds, 1)
	} else {
		m.nexus.Spec.LivenessProbe = DefaultProbe.DeepCopy()
	}

	// SuccessThreshold for Liveness Probes must be 1
	m.nexus.Spec.LivenessProbe.SuccessThreshold = 1

	if m.nexus.Spec.ReadinessProbe != nil {
		m.nexus.Spec.ReadinessProbe.FailureThreshold =
			util.EnsureMinimum(m.nexus.Spec.ReadinessProbe.FailureThreshold, 1)
		m.nexus.Spec.ReadinessProbe.InitialDelaySeconds =
			util.EnsureMinimum(m.nexus.Spec.ReadinessProbe.InitialDelaySeconds, 0)
		m.nexus.Spec.ReadinessProbe.PeriodSeconds =
			util.EnsureMinimum(m.nexus.Spec.ReadinessProbe.PeriodSeconds, 1)
		m.nexus.Spec.ReadinessProbe.SuccessThreshold =
			util.EnsureMinimum(m.nexus.Spec.ReadinessProbe.SuccessThreshold, 1)
		m.nexus.Spec.ReadinessProbe.TimeoutSeconds =
			util.EnsureMinimum(m.nexus.Spec.ReadinessProbe.TimeoutSeconds, 1)
	} else {
		m.nexus.Spec.ReadinessProbe = DefaultProbe.DeepCopy()
	}
}

// GetRequiredResources returns the resources initialized by the manager
func (m *Manager) GetRequiredResources() ([]resource.KubernetesResource, error) {
	if m.nexus == nil || m.client == nil {
		return nil, fmt.Errorf(mgrNotInit)
	}

	log.Debugf("Creating Deployment (%s)", m.nexus.Name)
	deployment := newDeployment(m.nexus)
	log.Debugf("Creating Service (%s)", m.nexus.Name)
	svc := newService(m.nexus)
	return []resource.KubernetesResource{deployment, svc}, nil
}

// GetDeployedResources returns the deployment-related resources deployed on the cluster
func (m *Manager) GetDeployedResources() ([]resource.KubernetesResource, error) {
	if m.nexus == nil || m.client == nil {
		return nil, fmt.Errorf(mgrNotInit)
	}

	var resources []resource.KubernetesResource
	if deployment, err := m.getDeployedDeployment(); err == nil {
		resources = append(resources, deployment)
	} else if !errors.IsNotFound(err) {
		log.Errorf("Could not fetch Deployment (%s): %v", m.nexus.Name, err)
		return nil, fmt.Errorf("could not fetch service (%s): %v", m.nexus.Name, err)
	}
	if service, err := m.getDeployedService(); err == nil {
		resources = append(resources, service)
	} else if !errors.IsNotFound(err) {
		log.Errorf("Could not fetch Service (%s): %v", m.nexus.Name, err)
		return nil, fmt.Errorf("could not fetch service (%s): %v", m.nexus.Name, err)
	}
	return resources, nil
}

func (m *Manager) getDeployedDeployment() (resource.KubernetesResource, error) {
	dep := &appsv1.Deployment{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	log.Debugf("Attempting to fetch deployed Deployment (%s)", m.nexus.Name)
	err := m.client.Get(ctx.TODO(), key, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed Service (%s)", m.nexus.Name)
		}
		return nil, err
	}
	return dep, nil
}

func (m *Manager) getDeployedService() (resource.KubernetesResource, error) {
	svc := &corev1.Service{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	log.Debugf("Attempting to fetch deployed Service (%s)", m.nexus.Name)
	err := m.client.Get(ctx.TODO(), key, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed Service (%s)", m.nexus.Name)
		}
		return nil, err
	}
	return svc, nil
}

// GetCustomComparator returns the custom comp function used to compare a deployment-related resource
// Returns nil if there is none
func (m *Manager) GetCustomComparator(t reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	// As Deployments and Services have default comparators we just return nil here
	return nil
}

// GetCustomComparators returns all custom comp functions in a map indexed by the resource type
// Returns nil if there are none
func (m *Manager) GetCustomComparators() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	// As Deployments and Services have default comparators we just return nil here
	return nil
}
