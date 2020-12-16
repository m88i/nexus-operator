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

package v1alpha1

import (
	"fmt"

	"github.com/m88i/nexus-operator/pkg/cluster/discovery"
	"github.com/m88i/nexus-operator/pkg/logger"
)

// validator uses its state to check if a given Nexus CR is valid
type validator struct {
	nexus                            *Nexus
	log                              logger.Logger
	routeAvailable, ingressAvailable bool
}

// newValidator builds a *validator with all necessary state. If it fails to construct this state, it returns an error.
// If it fails to construct this state, it assumes safe values and logs problems.
func newValidator(nexus *Nexus) *validator {
	log := logger.GetLoggerWithResource("admission_validation", nexus)

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

	return &validator{
		nexus:            nexus,
		log:              log,
		routeAvailable:   routeAvailable,
		ingressAvailable: ingressAvailable || legacyIngressAvailable,
	}
}

// validate returns an error if the Nexus CR is invalid
func (v *validator) validate() error {
	return v.validateNetworking()
}

func (v *validator) validateNetworking() error {
	if !v.nexus.Spec.Networking.Expose {
		v.log.Debug("'spec.networking.expose' set to 'false', ignoring networking configuration")
		return nil
	}

	if !v.ingressAvailable && v.nexus.Spec.Networking.ExposeAs == IngressExposeType {
		v.log.Warn("Ingresses are not available on your cluster. Make sure to be running Kubernetes > 1.14 or if you're running Openshift set ", "spec.networking.exposeAs", IngressExposeType, "Also try", NodePortExposeType)
		return fmt.Errorf("ingress expose required, but unavailable")
	}

	if !v.routeAvailable && v.nexus.Spec.Networking.ExposeAs == RouteExposeType {
		v.log.Warn("Routes are not available on your cluster. If you're running Kubernetes 1.14 or higher try setting ", "'spec.networking.exposeAs'", IngressExposeType, "Also try", NodePortExposeType)
		return fmt.Errorf("route expose required, but unavailable")
	}

	if v.nexus.Spec.Networking.ExposeAs == NodePortExposeType && v.nexus.Spec.Networking.NodePort == 0 {
		v.log.Warn("NodePort networking requires a port. Check the Nexus resource 'spec.networking.nodePort' parameter")
		return fmt.Errorf("nodeport expose required, but no port informed")
	}

	if v.nexus.Spec.Networking.ExposeAs == IngressExposeType && len(v.nexus.Spec.Networking.Host) == 0 {
		v.log.Warn("Ingress networking requires a host. Check the Nexus resource 'spec.networking.host' parameter")
		return fmt.Errorf("ingress expose required, but no host informed")
	}

	if len(v.nexus.Spec.Networking.TLS.SecretName) > 0 && v.nexus.Spec.Networking.ExposeAs != IngressExposeType {
		v.log.Warn("'spec.networking.tls.secretName' is only available when using an Ingress. Try setting ", "spec.networking.exposeAs'", IngressExposeType)
		return fmt.Errorf("tls secret name informed, but using route")
	}

	if v.nexus.Spec.Networking.TLS.Mandatory && v.nexus.Spec.Networking.ExposeAs != RouteExposeType {
		v.log.Warn("'spec.networking.tls.mandatory' is only available when using a Route. Try setting ", "spec.networking.exposeAs'", RouteExposeType)
		return fmt.Errorf("tls set to mandatory, but using ingress")
	}

	return nil
}
