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

	// Number of pod replicas desired. Defaults to 0.
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Replicas"
	Replicas int32 `json:"replicas"`

	// Full image tag name for this specific deployment. Will be ignored if `spec.useRedHatImage` is set to `true`.
	// Default: docker.io/sonatype/nexus3:latest
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes:image"
	Image string `json:"image,omitempty"`

	// Automatic updates configuration
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Automatic Update"
	AutomaticUpdate NexusAutomaticUpdate `json:"automaticUpdate,omitempty"`

	// The image pull policy for the Nexus image. If left blank behavior will be determined by the image tag (`Always` if "latest" and `IfNotPresent` otherwise).
	// Possible values: `Always`, `IfNotPresent` or `Never`.
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Image Pull Policy"
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

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

	// If you have access to Red Hat Container Catalog, set this to `true` to use the certified image provided by Sonatype
	// Defaults to `false`
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Use Red Hat Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	UseRedHatImage bool `json:"useRedHatImage"`

	// GenerateRandomAdminPassword enables the random password generation.
	// Defaults to `false`: the default password for a newly created instance is 'admin123', which should be changed in the first login.
	// If set to `true`, you must use the automatically generated 'admin' password, stored in the container's file system at `/nexus-data/admin.password`.
	// The operator uses the default credentials to create a user for itself to create default repositories.
	// If set to `true`, the repositories won't be created since the operator won't fetch for the random password.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Generate Random Admin Password"
	// +optional
	GenerateRandomAdminPassword bool `json:"generateRandomAdminPassword,omitempty"`

	// Networking definition
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=false
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Networking"
	Networking NexusNetworking `json:"networking,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount used to run the Pods. If left blank, a default ServiceAccount is created with the same name as the Nexus CR (`metadata.name`).
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

	// ServerOperations describes the options for the operations performed on the deployed server instance
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +optional
	ServerOperations ServerOperationsOpts `json:"serverOperations,omitempty"`
}

// NexusPersistence is the structure for the data persistent
// +k8s:openapi-gen=true
type NexusPersistence struct {
	// Flag to indicate if this instance installation will be persistent or not. If set to true a PVC is created for it.
	Persistent bool `json:"persistent"`
	// If persistent, the size of the Volume.
	// Defaults: 10Gi
	VolumeSize string `json:"volumeSize,omitempty"`
	// StorageClass used by the managed PVC.
	StorageClass string `json:"storageClass,omitempty"`
	// ExtraVolumes which should be mounted when deploying Nexus.
	// Updating this may lead to temporary unavailability while the new deployment with new volumes rolls out.
	// +optional
	ExtraVolumes []NexusVolume `json:"extraVolumes,omitempty"`
}

// NexusVolume embeds a Volume structure to represent a volume to be mounted in the Nexus pod at the specified MountPath
type NexusVolume struct {
	corev1.Volume `json:",inline"`
	// MountPath is the path where this volume should be mounted
	MountPath string `json:"mountPath"`
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
	// Annotations that should be added to the Ingress/Route resource
	// +optional
	// +nullable
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels that should be added to the Ingress/Route resource
	// +optional
	// +nullable
	Labels map[string]string `json:"labels,omitempty"`
	// Set to `true` to expose the Nexus application. Defaults to `false`.
	Expose bool `json:"expose,omitempty"`
	// Type of networking exposure: NodePort, Route or Ingress. Defaults to Route on OpenShift and Ingress on Kubernetes.
	// Routes are only available on Openshift and Ingresses are only available on Kubernetes.
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
	// Number of seconds after the container has started before probes are initiated.
	// Defaults to 240 seconds. Minimum value is 0.
	// +optional
	// +kubebuilder:validation:Minimum=0
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty" protobuf:"varint,2,opt,name=initialDelaySeconds"`
	// Number of seconds after which the probe times out.
	// Defaults to 15 seconds. Minimum value is 1.
	// +optional
	// +kubebuilder:validation:Minimum=1
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty" protobuf:"varint,3,opt,name=timeoutSeconds"`
	// How often (in seconds) to perform the probe.
	// Defaults to 10 seconds. Minimum value is 1.
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
	// When exposing via Route, set to `true` to only allow encrypted traffic using TLS (disables HTTP in favor of HTTPS). Defaults to `false`.
	// +optional
	Mandatory bool `json:"mandatory,omitempty"`
	// When exposing via Ingress, inform the name of the TLS secret containing certificate and private key for TLS encryption. It must be present in the same namespace as the Operator.
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// ServerOperationsOpts describes the options for the operations performed in the Nexus server deployed instance
type ServerOperationsOpts struct {
	// DisableRepositoryCreation disables the auto-creation of Apache, JBoss and Red Hat repositories and their addition to
	// the Maven Public group in this Nexus instance.
	// Defaults to `false` (always try to create the repos). Set this to `true` to not create them. Only works if `spec.generateRandomAdminPassword` is `false`.
	DisableRepositoryCreation bool `json:"disableRepositoryCreation,omitempty"`
	// DisableOperatorUserCreation disables the auto-creation of the `nexus-operator` user on the deployed server. This user performs
	// all the operations on the server (such as creating the community repos). If disabled, the Operator will use the default `admin` user.
	// Defaults to `false` (always create the user). Setting this to `true` is not recommended as it grants the Operator more privileges than it needs and it would not be possible to tell apart operations performed by the `admin` and the Operator.
	DisableOperatorUserCreation bool `json:"disableOperatorUserCreation,omitempty"`
}

// NexusAutomaticUpdate defines configuration for automatic updates
type NexusAutomaticUpdate struct {
	// Whether or not the Operator should perform automatic updates. Defaults to `false` (auto updates are enabled).
	// Is set to `false` if `spec.image` is not empty and is different from the default community image.
	// +optional
	Disabled bool `json:"disabled,omitempty"`
	// The Nexus image minor version the deployment should stay in. If left blank and automatic updates are enabled the latest minor is set.
	// +kubebuilder:validation:Minimum=0
	// +optional
	MinorVersion *int `json:"minorVersion,omitempty"` // must keep a pointer to tell apart uninformed from 0
}

// NexusStatus defines the observed state of Nexus
// +k8s:openapi-gen=true
type NexusStatus struct {
	// Condition status for the Nexus deployment
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="appsv1.DeploymentStatus"
	DeploymentStatus v1.DeploymentStatus `json:"deploymentStatus,omitempty"`
	// Will be "OK" when this Nexus instance is up
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	NexusStatus NexusStatusType `json:"nexusStatus,omitempty"`
	// Gives more information about a failure status
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Reason string `json:"reason,omitempty"`
	// Route for external service access
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	NexusRoute string `json:"nexusRoute,omitempty"`
	// Conditions reached during an update
	// +listType=atomic
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Update Conditions"
	UpdateConditions []string `json:"updateConditions,omitempty"`
	// ServerOperationsStatus describes the general status for the operations performed in the Nexus server instance
	ServerOperationsStatus OperationsStatus `json:"serverOperationsStatus,omitempty"`
}

// OperationsStatus describes the status for each operation made by the operator in the deployed Nexus Server
type OperationsStatus struct {
	ServerReady                  bool   `json:"serverReady,omitempty"`
	OperatorUserCreated          bool   `json:"operatorUserCreated,omitempty"`
	CommunityRepositoriesCreated bool   `json:"communityRepositoriesCreated,omitempty"`
	MavenCentralUpdated          bool   `json:"mavenCentralUpdated,omitempty"`
	Reason                       string `json:"reason,omitempty"`
	MavenPublicURL               string `json:"mavenPublicURL,omitempty"`
}

type NexusStatusType string

const (
	// NexusStatusOK is the ok status
	NexusStatusOK NexusStatusType = "OK"
	// NexusStatusFailure is the failed status
	NexusStatusFailure NexusStatusType = "Failure"
	// NexusStatusPending is the failed status
	NexusStatusPending NexusStatusType = "Pending"
)

// Nexus custom resource to deploy the Nexus Server
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=nexus,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Expose As",type="string",JSONPath=".spec.networking.exposeAs",description="Type of networking access"
// +kubebuilder:printcolumn:name="Update Disabled",type="boolean",JSONPath=".spec.automaticUpdate.disabled",description="Flag that indicates if automatic updates are disabled or not"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.nexusStatus",description="Instance Status"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.reason",description="Status reason"
// +kubebuilder:printcolumn:name="Maven Public URL",type="string",JSONPath=".status.serverOperationsStatus.mavenPublicURL",description="Internal Group Maven Public URL"
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

// NexusList contains a list of Nexus
// +kubebuilder:object:root=true
type NexusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nexus `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Nexus{}, &NexusList{})
}
