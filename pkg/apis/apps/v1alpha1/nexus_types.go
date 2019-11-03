package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NexusSpec defines the desired state of Nexus
// +k8s:openapi-gen=true
type NexusSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=1
	// Replicas is a number of Nexus pod replicas
	Replicas int16 `json:"replicas"`

	// +kubebuilder:validation:MinLength=5
	// Image is the full image tag name for this specific deployment
	Image string `json:"image"`

	// +optional
	Resources NexusResources `json:"resources,omitempty"`

	Credentials NexusCredentials `json:"credentials"`

	Persistence NexusPersistence `json:"persistence"`
}

// NexusResources is the request and limit definitions for the deployed pods
// +k8s:openapi-gen=true
type NexusResources struct {
	// +optional
	CPURequest string `json:"cpuRequest,omitempty"`
	// +optional
	CPULimit string `json:"cpuLimit,omitempty"`
	// +optional
	MemoryRequest string `json:"memoryRequest,omitempty"`
	// +optional
	MemoryLimit string `json:"memoryLimit,omitempty"`
}

// NexusCredentials is the credentials for the administrator user in the Nexus web console
// +k8s:openapi-gen=true
type NexusCredentials struct {
	// +kubebuilder:validation:MinLength=5
	User string `json:"user"`
	// +kubebuilder:validation:MinLength=5
	Password string `json:"password"`
}

// NexusPersistence is the structure for the data persistent
// +k8s:openapi-gen=true
type NexusPersistence struct {
	Persistent bool   `json:"persistent"`
	VolumeSize string `json:"volumeSize,omitempty"`
}

// NexusStatus defines the observed state of Nexus
// +k8s:openapi-gen=true
type NexusStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Condition is the general condition status for the Nexus deployment
	Condition string `json:"condition,omitempty"`
	// Host is Nexus Web Console URL
	Host string `json:"host,omitempty"`
	// AdminUser is the administrator username
	AdminUser string `json:"adminUser,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Nexus is the Schema for the nexus API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=nexus,scope=Namespaced
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
