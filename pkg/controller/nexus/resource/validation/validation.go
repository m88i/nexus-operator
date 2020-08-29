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

package validation

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"strings"

	"github.com/m88i/nexus-operator/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/update"
	"github.com/m88i/nexus-operator/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
)

const (
	discOCPFailureFormat      = "unable to determine if cluster is Openshift: %v"
	discFailureFormat         = "unable to determine if %s are available: %v" // resource type, error
	unspecifiedExposeAsFormat = "'spec.exposeAs' left unspecified, setting it to %s"
)

var log = logger.GetLogger("Nexus_validation")

type Validator struct {
	client client.Client
	scheme *runtime.Scheme

	routeAvailable, ingressAvailable, ocp bool
}

// NewValidator creates a new validator to set defaults, validate and update the Nexus CR
func NewValidator(client client.Client, scheme *runtime.Scheme, disc discovery.DiscoveryInterface) (*Validator, error) {
	routeAvailable, err := openshift.IsRouteAvailable(disc)
	if err != nil {
		return nil, fmt.Errorf(discFailureFormat, "routes", err)
	}

	ingressAvailable, err := kubernetes.IsIngressAvailable(disc)
	if err != nil {
		return nil, fmt.Errorf(discFailureFormat, "ingresses", err)
	}

	ocp, err := openshift.IsOpenShift(disc)
	if err != nil {
		return nil, fmt.Errorf(discOCPFailureFormat, err)
	}

	return &Validator{
		client:           client,
		scheme:           scheme,
		routeAvailable:   routeAvailable,
		ingressAvailable: ingressAvailable,
		ocp:              ocp,
	}, nil
}

// SetDefaultsAndValidate returns a copy of the parameter Nexus with defaults set and an error if validation fails.
func (v *Validator) SetDefaultsAndValidate(nexus *v1alpha1.Nexus) (*v1alpha1.Nexus, error) {
	n := v.setDefaults(nexus)
	return n, v.validate(n)
}

func (v *Validator) validate(nexus *v1alpha1.Nexus) error {
	return v.validateNetworking(nexus)
}

func (v *Validator) validateNetworking(nexus *v1alpha1.Nexus) error {
	if !nexus.Spec.Networking.Expose {
		log.Debugf("'spec.networking.expose' set to 'false', ignoring networking configuration")
		return nil
	}

	if !v.ingressAvailable && nexus.Spec.Networking.ExposeAs == v1alpha1.IngressExposeType {
		log.Errorf("Ingresses are not available on your cluster. Make sure to be running Kubernetes > 1.14 or if you're running Openshift set 'spec.networking.exposeAs' to '%s'. Alternatively you may also try '%s'", v1alpha1.IngressExposeType, v1alpha1.NodePortExposeType)
		return fmt.Errorf("ingress expose required, but unavailable")
	}

	if !v.routeAvailable && nexus.Spec.Networking.ExposeAs == v1alpha1.RouteExposeType {
		log.Errorf("Routes are not available on your cluster. If you're running Kubernetes 1.14 or higher try setting 'spec.networking.exposeAs' to '%s'. Alternatively you may also try '%s'", v1alpha1.IngressExposeType, v1alpha1.NodePortExposeType)
		return fmt.Errorf("route expose required, but unavailable")
	}

	if nexus.Spec.Networking.ExposeAs == v1alpha1.NodePortExposeType && nexus.Spec.Networking.NodePort == 0 {
		log.Errorf("NodePort networking requires a port. Check the Nexus resource 'spec.networking.nodePort' parameter")
		return fmt.Errorf("nodeport expose required, but no port informed")
	}

	if nexus.Spec.Networking.ExposeAs == v1alpha1.IngressExposeType && len(nexus.Spec.Networking.Host) == 0 {
		log.Errorf("Ingress networking requires a host. Check the Nexus resource 'spec.networking.host' parameter")
		return fmt.Errorf("ingress expose required, but no host informed")
	}

	if len(nexus.Spec.Networking.TLS.SecretName) > 0 && nexus.Spec.Networking.ExposeAs != v1alpha1.IngressExposeType {
		log.Errorf("'spec.networking.tls.secretName' is only available when using an Ingress. Try setting 'spec.networking.exposeAs' to '%s'", v1alpha1.IngressExposeType)
		return fmt.Errorf("tls secret name informed, but using route")
	}

	if nexus.Spec.Networking.TLS.Mandatory && nexus.Spec.Networking.ExposeAs != v1alpha1.RouteExposeType {
		log.Errorf("'spec.networking.tls.mandatory' is only available when using a Route. Try setting 'spec.networking.exposeAs' to '%s'", v1alpha1.RouteExposeType)
		return fmt.Errorf("tls set to mandatory, but using ingress")
	}

	return nil
}

func (v *Validator) setDefaults(nexus *v1alpha1.Nexus) *v1alpha1.Nexus {
	n := nexus.DeepCopy()
	v.setDeploymentDefaults(n)
	v.setUpdateDefaults(n)
	v.setNetworkingDefaults(n)
	v.setPersistenceDefaults(n)
	v.setSecurityDefaults(n)
	return n
}

func (v *Validator) setDeploymentDefaults(nexus *v1alpha1.Nexus) {
	v.setResourcesDefaults(nexus)
	v.setImageDefaults(nexus)
	v.setProbeDefaults(nexus)
}

func (v *Validator) setResourcesDefaults(nexus *v1alpha1.Nexus) {
	if nexus.Spec.Resources.Requests == nil && nexus.Spec.Resources.Limits == nil {
		nexus.Spec.Resources = DefaultResources
	}
}

func (v *Validator) setImageDefaults(nexus *v1alpha1.Nexus) {
	if nexus.Spec.UseRedHatImage {
		if len(nexus.Spec.Image) > 0 {
			log.Warnf("Nexus CR configured to the use Red Hat Certified Image, ignoring 'spec.image' field.")
		}
		nexus.Spec.Image = NexusCertifiedImage
	} else if len(nexus.Spec.Image) == 0 {
		nexus.Spec.Image = NexusCommunityImage
	}

	if len(nexus.Spec.ImagePullPolicy) > 0 &&
		nexus.Spec.ImagePullPolicy != corev1.PullAlways &&
		nexus.Spec.ImagePullPolicy != corev1.PullIfNotPresent &&
		nexus.Spec.ImagePullPolicy != corev1.PullNever {

		log.Warnf("Invalid 'spec.imagePullPolicy', unsetting the value. The pull policy will be determined by the image tag. Consider setting this value to '%s', '%s' or '%s'", corev1.PullAlways, corev1.PullIfNotPresent, corev1.PullNever)
		nexus.Spec.ImagePullPolicy = ""
	}
}

func (v *Validator) setProbeDefaults(nexus *v1alpha1.Nexus) {
	if nexus.Spec.LivenessProbe != nil {
		nexus.Spec.LivenessProbe.FailureThreshold =
			ensureMinimum(nexus.Spec.LivenessProbe.FailureThreshold, 1)
		nexus.Spec.LivenessProbe.InitialDelaySeconds =
			ensureMinimum(nexus.Spec.LivenessProbe.InitialDelaySeconds, 0)
		nexus.Spec.LivenessProbe.PeriodSeconds =
			ensureMinimum(nexus.Spec.LivenessProbe.PeriodSeconds, 1)
		nexus.Spec.LivenessProbe.TimeoutSeconds =
			ensureMinimum(nexus.Spec.LivenessProbe.TimeoutSeconds, 1)
	} else {
		nexus.Spec.LivenessProbe = DefaultProbe.DeepCopy()
	}

	// SuccessThreshold for Liveness Probes must be 1
	nexus.Spec.LivenessProbe.SuccessThreshold = 1

	if nexus.Spec.ReadinessProbe != nil {
		nexus.Spec.ReadinessProbe.FailureThreshold =
			ensureMinimum(nexus.Spec.ReadinessProbe.FailureThreshold, 1)
		nexus.Spec.ReadinessProbe.InitialDelaySeconds =
			ensureMinimum(nexus.Spec.ReadinessProbe.InitialDelaySeconds, 0)
		nexus.Spec.ReadinessProbe.PeriodSeconds =
			ensureMinimum(nexus.Spec.ReadinessProbe.PeriodSeconds, 1)
		nexus.Spec.ReadinessProbe.SuccessThreshold =
			ensureMinimum(nexus.Spec.ReadinessProbe.SuccessThreshold, 1)
		nexus.Spec.ReadinessProbe.TimeoutSeconds =
			ensureMinimum(nexus.Spec.ReadinessProbe.TimeoutSeconds, 1)
	} else {
		nexus.Spec.ReadinessProbe = DefaultProbe.DeepCopy()
	}
}

// must be called only after image defaults have been set
func (v *Validator) setUpdateDefaults(nexus *v1alpha1.Nexus) {
	if nexus.Spec.AutomaticUpdate.Disabled {
		return
	}

	image := strings.Split(nexus.Spec.Image, ":")[0]
	if image != NexusCommunityImage {
		log.Warnf("Automatic Updates are enabled, but 'spec.image' is not using the community image (%s). Disabling automatic updates", NexusCommunityImage)
		nexus.Spec.AutomaticUpdate.Disabled = true
		return
	}

	if nexus.Spec.AutomaticUpdate.MinorVersion == nil {
		log.Debugf("Automatic Updates are enabled, but no minor was informed. Fetching the most recent...")
		minor, err := update.GetLatestMinor()
		if err != nil {
			log.Errorf("Unable to fetch the most recent minor: %v. Disabling automatic updates.", err)
			nexus.Spec.AutomaticUpdate.Disabled = true
			createChangedNexusEvent(nexus, v.scheme, v.client, "spec.automaticUpdate.disabled")
			return
		}
		nexus.Spec.AutomaticUpdate.MinorVersion = &minor
	}

	log.Debugf("Fetching the latest micro from minor %d", *nexus.Spec.AutomaticUpdate.MinorVersion)
	tag, ok := update.GetLatestMicro(*nexus.Spec.AutomaticUpdate.MinorVersion)
	if !ok {
		// the informed minor doesn't exist, let's try the latest minor
		log.Warnf("Latest tag for minor version (%d) not found. Trying the latest minor instead", *nexus.Spec.AutomaticUpdate.MinorVersion)
		minor, err := update.GetLatestMinor()
		if err != nil {
			log.Errorf("Unable to fetch the most recent minor: %v. Disabling automatic updates.", err)
			nexus.Spec.AutomaticUpdate.Disabled = true
			createChangedNexusEvent(nexus, v.scheme, v.client, "spec.automaticUpdate.disabled")
			return
		}
		log.Infof("Setting 'spec.automaticUpdate.minorVersion to %d", minor)
		nexus.Spec.AutomaticUpdate.MinorVersion = &minor
		// no need to check for the tag existence here,
		// we would have gotten an error from GetLatestMinor() if it didn't
		tag, _ = update.GetLatestMicro(minor)
	}
	newImage := fmt.Sprintf("%s:%s", image, tag)
	log.Debugf("Replacing 'spec.image' (%s) with '%s'", nexus.Spec.Image, newImage)
	nexus.Spec.Image = newImage
}

func (v *Validator) setNetworkingDefaults(nexus *v1alpha1.Nexus) {
	if !nexus.Spec.Networking.Expose {
		return
	}

	if len(nexus.Spec.Networking.ExposeAs) == 0 {
		if v.ocp {
			log.Infof(unspecifiedExposeAsFormat, v1alpha1.RouteExposeType)
			nexus.Spec.Networking.ExposeAs = v1alpha1.RouteExposeType
		} else if v.ingressAvailable {
			log.Infof(unspecifiedExposeAsFormat, v1alpha1.IngressExposeType)
			nexus.Spec.Networking.ExposeAs = v1alpha1.IngressExposeType
		} else {
			// we're on kubernetes < 1.14
			// try setting nodePort, validation will catch it if impossible
			log.Info("On Kubernetes, but Ingresses are not available")
			log.Infof(unspecifiedExposeAsFormat, v1alpha1.NodePortExposeType)
			nexus.Spec.Networking.ExposeAs = v1alpha1.NodePortExposeType
		}
	}
}

func (v *Validator) setPersistenceDefaults(nexus *v1alpha1.Nexus) {
	if nexus.Spec.Persistence.Persistent && len(nexus.Spec.Persistence.VolumeSize) == 0 {
		nexus.Spec.Persistence.VolumeSize = DefaultVolumeSize
	}
}

func (v *Validator) setSecurityDefaults(nexus *v1alpha1.Nexus) {
	if len(nexus.Spec.ServiceAccountName) == 0 {
		nexus.Spec.ServiceAccountName = nexus.Name
	}
}

func ensureMinimum(value, minimum int32) int32 {
	if value < minimum {
		return minimum
	}
	return value
}
