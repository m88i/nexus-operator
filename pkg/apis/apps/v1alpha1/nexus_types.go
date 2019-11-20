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
	// Default: 1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=1
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Replicas"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.description="Replicas"
	Replicas int32 `json:"replicas"`

	// Full image tag name for this specific deployment
	// Default: docker.io/sonatype/nexus3:latest
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.description="Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes:image"
	Image string `json:"image,omitempty"`

	// Defined Resources for the Nexus instance
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Resources"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.description="Resources"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Persistence definition
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=false
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Persistence"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.description="Persistence"
	Persistence NexusPersistence `json:"persistence"`

	// If you have access to Red Hat Container Catalog, turn this to true to use the certified image provided by Sonatype
	// Default: false
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Use Red Hat Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.description="Use Red Hat Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	UseRedHatImage bool `json:"useRedHatImage,omitempty"`
}

// NexusPersistence is the structure for the data persistent
// +k8s:openapi-gen=true
type NexusPersistence struct {
	// Flag to indicate if this instance will be persistent or not
	Persistent bool `json:"persistent"`
	// If persistent, the size of the Volume.
	// Defaults: 10Gi
	VolumeSize string `json:"volumeSize,omitempty"`
}

// NexusStatus defines the observed state of Nexus
// +k8s:openapi-gen=true
type NexusStatus struct {
	// Condition status for the Nexus deployment
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="appsv1.DeploymentStatus"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.description="appsv1.DeploymentStatus"
	DeploymentStatus v1.DeploymentStatus `json:"deploymentStatus,omitempty"`
	// Will be "OK" when all objects are created successfully
	NexusStatus string `json:"nexusStatus,omitempty"`
	// Route for external service access
	NexusRoute string `json:"nexusRoute,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Nexus is the Schema for the nexus API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
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
