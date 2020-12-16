package v1alpha1

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/m88i/nexus-operator/pkg/cluster/discovery"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/m88i/nexus-operator/pkg/logger"
)

const (
	unspecifiedExposeAsFormat = "'spec.exposeAs' left unspecified, setting to: "
)

// mutator uses its state to know what changes to apply to a Nexus CR
type mutator struct {
	nexus                            *Nexus
	log                              logger.Logger
	routeAvailable, ingressAvailable bool
}

// newMutator builds a *mutator with all necessary state.
// If it fails to construct this state, it assumes safe values and logs problems.
func newMutator(nexus *Nexus) *mutator {
	log := logger.GetLoggerWithResource("admission_mutation", nexus)

	routeAvailable, err := discovery.IsRouteAvailable()
	if err != nil {
		log.Error(err, "Unable to determine if Routes are available. Will assume they're not.")
	}

	ingressAvailable, err := discovery.IsIngressAvailable()
	if err != nil {
		log.Error(err, "Unable to determine if v1 Ingresses are available. Will assume they're not.")
	}

	legacyIngressAvailable, err := discovery.IsLegacyIngressAvailable()
	if err != nil {
		log.Error(err, "Unable to determine if v1beta1 Ingresses are available. Will assume they're not.")
	}

	return &mutator{
		nexus:          nexus,
		log:            log,
		routeAvailable: routeAvailable,
		// TODO: create a IsAnyIngressAvailable function in discovery, we don't care about which one in admission
		ingressAvailable: ingressAvailable || legacyIngressAvailable,
	}
}

// mutate make all necessary changes to the Nexus resource prior to validation
func (m *mutator) mutate() {
	m.mutateDeployment()
	m.mutateAutomaticUpdate()
	m.mutateNetworking()
	m.mutatePersistence()
	m.mutateSecurity()
}

func (m *mutator) mutateDeployment() {
	m.mutateReplicas()
	m.mutateResources()
	m.mutateImage()
	m.mutateProbe()
}

func (m *mutator) mutateReplicas() {
	if m.nexus.Spec.Replicas > maxReplicas {
		m.log.Warn("Number of replicas not supported", "MaxSupportedReplicas", maxReplicas, "DesiredReplicas", m.nexus.Spec.Replicas)
		m.nexus.Spec.Replicas = maxReplicas
	}
}

func (m *mutator) mutateResources() {
	if m.nexus.Spec.Resources.Requests == nil && m.nexus.Spec.Resources.Limits == nil {
		m.nexus.Spec.Resources = DefaultResources
	}
}

func (m *mutator) mutateImage() {
	if m.nexus.Spec.UseRedHatImage {
		if len(m.nexus.Spec.Image) > 0 {
			m.log.Warn("Nexus CR configured to the use Red Hat Certified Image, ignoring 'spec.image' field.")
		}
		m.nexus.Spec.Image = NexusCertifiedImage
	}
	if len(m.nexus.Spec.Image) == 0 {
		m.nexus.Spec.Image = NexusCommunityImage
	}

	if len(m.nexus.Spec.ImagePullPolicy) > 0 &&
		m.nexus.Spec.ImagePullPolicy != corev1.PullAlways &&
		m.nexus.Spec.ImagePullPolicy != corev1.PullIfNotPresent &&
		m.nexus.Spec.ImagePullPolicy != corev1.PullNever {

		m.log.Warn("Invalid 'spec.imagePullPolicy', unsetting the value. The pull policy will be determined by the image tag. Valid value are", "#1", corev1.PullAlways, "#2", corev1.PullIfNotPresent, "#3", corev1.PullNever)
		m.nexus.Spec.ImagePullPolicy = ""
	}
}

func (m *mutator) mutateProbe() {
	if m.nexus.Spec.LivenessProbe != nil {
		m.nexus.Spec.LivenessProbe.FailureThreshold =
			ensureMinimum(m.nexus.Spec.LivenessProbe.FailureThreshold, 1)
		m.nexus.Spec.LivenessProbe.InitialDelaySeconds =
			ensureMinimum(m.nexus.Spec.LivenessProbe.InitialDelaySeconds, 0)
		m.nexus.Spec.LivenessProbe.PeriodSeconds =
			ensureMinimum(m.nexus.Spec.LivenessProbe.PeriodSeconds, 1)
		m.nexus.Spec.LivenessProbe.TimeoutSeconds =
			ensureMinimum(m.nexus.Spec.LivenessProbe.TimeoutSeconds, 1)
	} else {
		m.nexus.Spec.LivenessProbe = DefaultProbe.DeepCopy()
	}

	// SuccessThreshold for Liveness Probes must be 1
	m.nexus.Spec.LivenessProbe.SuccessThreshold = 1

	if m.nexus.Spec.ReadinessProbe != nil {
		m.nexus.Spec.ReadinessProbe.FailureThreshold =
			ensureMinimum(m.nexus.Spec.ReadinessProbe.FailureThreshold, 1)
		m.nexus.Spec.ReadinessProbe.InitialDelaySeconds =
			ensureMinimum(m.nexus.Spec.ReadinessProbe.InitialDelaySeconds, 0)
		m.nexus.Spec.ReadinessProbe.PeriodSeconds =
			ensureMinimum(m.nexus.Spec.ReadinessProbe.PeriodSeconds, 1)
		m.nexus.Spec.ReadinessProbe.SuccessThreshold =
			ensureMinimum(m.nexus.Spec.ReadinessProbe.SuccessThreshold, 1)
		m.nexus.Spec.ReadinessProbe.TimeoutSeconds =
			ensureMinimum(m.nexus.Spec.ReadinessProbe.TimeoutSeconds, 1)
	} else {
		m.nexus.Spec.ReadinessProbe = DefaultProbe.DeepCopy()
	}
}

func (m *mutator) mutateAutomaticUpdate() {
	if m.nexus.Spec.AutomaticUpdate.Disabled {
		return
	}

	image := strings.Split(m.nexus.Spec.Image, ":")[0]
	if image != NexusCommunityImage {
		m.log.Warn("Automatic Updates are enabled, but 'spec.image' is not using the community image. Disabling automatic updates", "Community Image", NexusCommunityImage)
		m.nexus.Spec.AutomaticUpdate.Disabled = true
		return
	}

	if m.nexus.Spec.AutomaticUpdate.MinorVersion == nil {
		m.log.Debug("Automatic Updates are enabled, but no minor was informed. Fetching the most recent...")
		minor, err := framework.GetLatestMinor()
		if err != nil {
			m.log.Error(err, "Unable to fetch the most recent minor. Disabling automatic updates.")
			m.nexus.Spec.AutomaticUpdate.Disabled = true
			return
		}
		m.nexus.Spec.AutomaticUpdate.MinorVersion = &minor
	}

	m.log.Debug("Fetching the latest micro from minor", "MinorVersion", *m.nexus.Spec.AutomaticUpdate.MinorVersion)
	tag, ok := framework.GetLatestMicro(*m.nexus.Spec.AutomaticUpdate.MinorVersion)
	if !ok {
		// the informed minor doesn't exist, let's try the latest minor
		m.log.Warn("Latest tag for minor version not found. Trying the latest minor instead", "Informed tag", *m.nexus.Spec.AutomaticUpdate.MinorVersion)
		minor, err := framework.GetLatestMinor()
		if err != nil {
			m.log.Error(err, "Unable to fetch the most recent minor: %m. Disabling automatic updates.")
			m.nexus.Spec.AutomaticUpdate.Disabled = true
			return
		}
		m.log.Info("Setting 'spec.automaticUpdate.minorVersion to", "MinorTag", minor)
		m.nexus.Spec.AutomaticUpdate.MinorVersion = &minor
		// no need to check for the tag existence here,
		// we would have gotten an error from GetLatestMinor() if it didn't
		tag, _ = framework.GetLatestMicro(minor)
	}
	newImage := fmt.Sprintf("%s:%s", image, tag)
	if newImage != m.nexus.Spec.Image {
		m.log.Debug("Replacing 'spec.image'", "OldImage", m.nexus.Spec.Image, "NewImage", newImage)
		m.nexus.Spec.Image = newImage
	}
}

func (m *mutator) mutateNetworking() {
	if !m.nexus.Spec.Networking.Expose {
		return
	}

	if len(m.nexus.Spec.Networking.ExposeAs) == 0 {
		if m.routeAvailable {
			m.log.Info(unspecifiedExposeAsFormat, "ExposeType", RouteExposeType)
			m.nexus.Spec.Networking.ExposeAs = RouteExposeType
		} else if m.ingressAvailable {
			m.log.Info(unspecifiedExposeAsFormat, "ExposeType", IngressExposeType)
			m.nexus.Spec.Networking.ExposeAs = IngressExposeType
		} else {
			// we're on kubernetes < 1.14
			// try setting nodePort, validation will catch it if impossible
			m.log.Info("On Kubernetes, but Ingresses are not available")
			m.log.Info(unspecifiedExposeAsFormat, "ExposeType", NodePortExposeType)
			m.nexus.Spec.Networking.ExposeAs = NodePortExposeType
		}
	}
}

func (m *mutator) mutatePersistence() {
	if m.nexus.Spec.Persistence.Persistent && len(m.nexus.Spec.Persistence.VolumeSize) == 0 {
		m.nexus.Spec.Persistence.VolumeSize = DefaultVolumeSize
	}
}

func (m *mutator) mutateSecurity() {
	if len(m.nexus.Spec.ServiceAccountName) == 0 {
		m.nexus.Spec.ServiceAccountName = m.nexus.Name
	}
}

func ensureMinimum(value, minimum int32) int32 {
	if value < minimum {
		return minimum
	}
	return value
}
