/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// in api/v1alpha1/ofbiz_types.go

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OfbizSpec defines the desired state of Ofbiz
type OfbizSpec struct {
	// Number of desired pods. Defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Size int32 `json:"size"`

	// The container image for OFBiz.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Configuration for the external PostgreSQL database.
	// +kubebuilder:validation:Required
	Database DatabaseSpec `json:"database"`

	// Web configuration, including SSL.
	// +optional
	Web WebSpec `json:"web,omitempty"`

	// Configuration for initial admin account.
	// +optional
	InitialAdmin AdminAccountSpec `json:"initialAdmin,omitempty"`

	// Storage configuration for OFBiz data and logs.
	// +optional
	Storage StorageSpec `json:"storage,omitempty"`
}

// DatabaseSpec defines the external database connection details.
type DatabaseSpec struct {
	// Hostname of the PostgreSQL server.
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// Port of the PostgreSQL server.
	// +kubebuilder:default=5432
	Port int32 `json:"port"`

	// Name of the database.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Username for the database connection.
	// +kubebuilder:validation:Required
	User string `json:"user"`

	// Name of the Secret containing the database password. The secret must have a key named 'password'.
	// +kubebuilder:validation:Required
	PasswordSecretName string `json:"passwordSecretName"`
}

// WebSpec defines web server configuration.
type WebSpec struct {
	// Name of the Secret of type kubernetes.io/tls for SSL.
	// The secret must contain 'tls.crt' and 'tls.key'.
	// +optional
	SslSecretName string `json:"sslSecretName,omitempty"`
}

// AdminAccountSpec defines the initial admin credentials.
type AdminAccountSpec struct {
	// Name of the Secret containing the admin password. The secret must have a key named 'password'.
	// The operator will use this to set the initial admin password.
	// +optional
	PasswordSecretName string `json:"passwordSecretName,omitempty"`
}

// StorageSpec defines the persistence configuration.
type StorageSpec struct {
	// Name of an existing ConfigMap with OFBiz configuration files.
	// If not provided, the operator can create a default one.
	// +optional
	ConfigurationConfigMapName string `json:"configurationConfigMapName,omitempty"`

	// PersistentVolumeClaim configuration for runtime data.
	// +optional
	Persistence PersistenceSpec `json:"persistence,omitempty"`
}

// PersistenceSpec defines the PVC for stateful data.
type PersistenceSpec struct {
	// Enable persistence. Defaults to true.
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// Storage class to use for the PVC.
	// +optional
	StorageClassName string `json:"storageClassName,omitempty"`

	// Size of the persistent volume. e.g., "10Gi".
	// +kubebuilder:default="5Gi"
	Size resource.Quantity `json:"size"`
}

// OfbizStatus defines the observed state of Ofbiz
type OfbizStatus struct {
	// Represents the observations of a Ofbiz's current state.
	// Known .status.conditions.type are: "Available", "Progressing", and "Degraded"
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// List of pod names managed by this Ofbiz instance.
	Nodes []string `json:"nodes,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ofbiz is the Schema for the ofbizs API
type Ofbiz struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OfbizSpec   `json:"spec,omitempty"`
	Status OfbizStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OfbizList contains a list of Ofbiz
type OfbizList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ofbiz `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ofbiz{}, &OfbizList{})
}

/* originally scaffolded one
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OfbizSpec defines the desired state of Ofbiz.
type OfbizSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Ofbiz. Edit ofbiz_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// OfbizStatus defines the observed state of Ofbiz.
type OfbizStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Ofbiz is the Schema for the ofbizzes API.
type Ofbiz struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OfbizSpec   `json:"spec,omitempty"`
	Status OfbizStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OfbizList contains a list of Ofbiz.
type OfbizList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ofbiz `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ofbiz{}, &OfbizList{})
}

*/
