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

package v1alpha1

import (
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NexusSpec defines the desired state of Nexus
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=nexus,scope=Namespaced
type NexusSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Number of pods replicas desired
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Replicas"
	Replicas int32 `json:"replicas"`

	// Full image tag name for this specific deployment
	// Default: docker.io/sonatype/nexus3:latest
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes:image"
	Image string `json:"image,omitempty"`

	// Defined Resources for the Nexus instance
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Resources"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Persistence definition
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=false
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Persistence"
	Persistence NexusPersistence `json:"persistence"`

	// If you have access to Red Hat Container Catalog, turn this to true to use the certified image provided by Sonatype
	// Default: false
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Use Red Hat Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	UseRedHatImage bool `json:"useRedHatImage"`

	// Networking definition
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=false
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Networking"
	Networking NexusNetworking `json:"networking,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount used to run the Pods. If left blank, a default ServiceAccount is created with the same name as the Nexus CR.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Service Account"
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// LivenessProbe describes how the Nexus container liveness probe should work
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=false
	// +optional
	LivenessProbe *NexusProbe `json:"livenessProbe,omitempty"`

	// ReadinessProbe describes how the Nexus container readiness probe should work
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=false
	// +optional
	ReadinessProbe *NexusProbe `json:"readinessProbe,omitempty"`
}

// NexusPersistence is the structure for the data persistent
// +k8s:openapi-gen=true
type NexusPersistence struct {
	// Flag to indicate if this instance will be persistent or not
	Persistent bool `json:"persistent"`
	// If persistent, the size of the Volume.
	// Defaults: 10Gi
	VolumeSize string `json:"volumeSize,omitempty"`
	// StorageClass used by the managed PVC.
	StorageClass string `json:"storageClass,omitempty"`
}

// NexusNetworkingExposeType defines how to expose Nexus service
type NexusNetworkingExposeType string

const (
	// NodePortExposeType The service is exposed via NodePort
	NodePortExposeType NexusNetworkingExposeType = "NodePort"
	// RouteExposeType On OpenShift, the service is exposed via a custom Route
	RouteExposeType NexusNetworkingExposeType = "Route"
	// IngressExposeType Supported on Kubernetes only, the service is exposed via NGINX Ingress
	IngressExposeType NexusNetworkingExposeType = "Ingress"
)

// NexusNetworking is the base structure for Nexus networking information
type NexusNetworking struct {
	// Set to `true` to expose the Nexus application. Default to false.
	Expose bool `json:"expose,omitempty"`
	// Type of networking exposure: NodePort, Route or Ingress. Default to Route on OpenShift and Ingress on Kubernetes.
	// +kubebuilder:validation:Enum=NodePort;Route;Ingress
	ExposeAs NexusNetworkingExposeType `json:"exposeAs,omitempty"`
	// Host where the Nexus service is exposed. This attribute is required if the service is exposed via Ingress.
	Host string `json:"host,omitempty"`
	// NodePort defined in the exposed service. Required if exposed via NodePort.
	NodePort int32 `json:"nodePort,omitempty"`
	// TLS/SSL-related configuration
	// +optional
	TLS NexusNetworkingTLS `json:"tls,omitempty"`
}

// NexusProbe describes a health check to be performed against a container to determine whether it is
// alive or ready to receive traffic.
// +k8s:openapi-gen=true
type NexusProbe struct {
	// Number of seconds after the container has started before liveness probes are initiated.
	// +optional
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty" protobuf:"varint,2,opt,name=initialDelaySeconds"`
	// Number of seconds after which the probe times out.
	// Defaults to 1 second. Minimum value is 1.
	// +optional
	// +kubebuilder:validation:Minimum=1
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty" protobuf:"varint,3,opt,name=timeoutSeconds"`
	// How often (in seconds) to perform the probe.
	// Default to 10 seconds. Minimum value is 1.
	// +optional
	// +kubebuilder:validation:Minimum=1
	PeriodSeconds int32 `json:"periodSeconds,omitempty" protobuf:"varint,4,opt,name=periodSeconds"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
	// +optional
	// +kubebuilder:validation:Minimum=1
	SuccessThreshold int32 `json:"successThreshold,omitempty" protobuf:"varint,5,opt,name=successThreshold"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	// Defaults to 3. Minimum value is 1.
	// +optional
	// +kubebuilder:validation:Minimum=1
	FailureThreshold int32 `json:"failureThreshold,omitempty" protobuf:"varint,6,opt,name=failureThreshold"`
}

// NexusNetworkingTLS defines TLS/SSL-related configuration
type NexusNetworkingTLS struct {
	// When exposing via Route, set to `true` to only allow encrypted traffic using TLS (disables HTTP in favor of HTTPS). Defaults to false.
	// +optional
	Mandatory bool `json:"mandatory,omitempty"`
	// When exposing via Ingress, inform the name of the TLS secret containing certificate and private key for TLS encryption. It must be present in the same namespace as the Operator.
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// NexusStatus defines the observed state of Nexus
// +k8s:openapi-gen=true
type NexusStatus struct {
	// Condition status for the Nexus deployment
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="appsv1.DeploymentStatus"
	DeploymentStatus v1.DeploymentStatus `json:"deploymentStatus,omitempty"`
	// Will be "OK" when all objects are created successfully
	NexusStatus string `json:"nexusStatus,omitempty"`
	// Route for external service access
	NexusRoute string `json:"nexusRoute,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Nexus custom resource to deploy the Nexus Server
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=nexus,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Nexus"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployment,v1,\"A Kubernetes Deployment\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Service,v1,\"A Kubernetes Service\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="PersistentVolumeClaim,v1,\"A Kubernetes PersistentVolumeClaim\""
type Nexus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NexusSpec   `json:"spec,omitempty"`
	Status NexusStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NexusList contains a list of Nexus
type NexusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nexus `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Nexus{}, &NexusList{})
}
