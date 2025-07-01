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

	// The container image for OFBiz.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Number of desired pods. Defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Size int32 `json:"replicas"`

	// Configuration for the external PostgreSQL database.
	// +kubebuilder:validation:Required
	Database DatabaseSpec `json:"database"`

	// If provided, the operator will manage the database, user, and permissions.
	// +optional
	// PostgresAdmin *PostgresAdminSpec `json:"postgresAdmin,omitempty"`

	// Web configuration, including SSL.
	// +optional
	Web WebSpec `json:"web,omitempty"`

	// Configuration for initial admin account.
	// +optional
	InitialAdmin AdminAccountSpec `json:"initialAdmin,omitempty"`

	// Storage configuration for OFBiz data and logs.
	// +optional
	Storage StorageSpec `json:"storage,omitempty"`

	// Java Opts.
	// +kubebuilder:default:"-Xms512m -Xmx1024m"
	JavaOpts string `json:"javaOpts,omitempty"`

	// AJP port enablement.
	// +kubebuilder:default:false
	EnableAJP bool `json:"enableAJP,omitempty"`
}

// PasswordSource defines a source for a password.
// Only one of its fields may be set.
// +kubebuilder:validation:XValidation:rule="has(self.value) != has(self.secretName)", message="exactly one of `value` or `secretName` must be specified"
type PasswordSource struct {
	// Value of the password.
	// The operator will create a Secret to store this value.
	// +optional
	Value string `json:"value,omitempty"`
	// Name of an existing Secret. The Secret must have a key named 'password'.
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// ConfigurationSource defines a source for OFBiz configuration.
// Only one of its fields may be set.
// +kubebuilder:validation:XValidation:rule="has(self.entityEngineXML) != has(self.configMapName)", message="exactly one of `entityEngineXML` or `configMapName` must be specified"
type ConfigurationSource struct {
	// Inline entityengine.xml content.
	// The operator will create a ConfigMap to store this.
	// +optional
	EntityEngineXML string `json:"entityEngineXML,omitempty"`
	// Name of an existing ConfigMap with OFBiz configuration files.
	// +optional
	ConfigMapName string `json:"configMapName,omitempty"`
}

// DatabaseSpec defines the external database connection details.
type DatabaseSpec struct {
	// If provided, the operator will manage the database, user, and permissions.
	// +optional
	PostgresAdmin *PostgresAdminSpec `json:"postgresAdmin,omitempty"`
	// Automatically comissioned if postgresAdmin specified, else should be manually created beforehand.

	// Hostname of the PostgreSQL server to connect to as an administrator.
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// Port of the PostgreSQL server. Defaults to 5432.
	// +kubebuilder:default=5432
	Port int32 `json:"port"`

	// SSL Mode for the admin connection (e.g., 'disable', 'require', 'verify-full').
	// +kubebuilder:default=disable
	SslMode string `json:"sslMode"`

	// Auto create ofbiz database
	// +kubebuilder:default=false
	CreateDB bool `json:"createDB,omitempty"`

	// ofbiz database
	OfbizDB OfbizDBSpec `json:"ofbizDB,omitempty"`

	// ofbiz OLAP database
	OfbizOLAPDB OfbizDBSpec `json:"ofbizOLAPDB,omitempty"`

	// ofbiz Tenant database
	OfbizTenantDB OfbizDBSpec `json:"ofbizTenantDB,omitempty"`

	// Data Load Mode (e.g., 'none', 'seed', 'demo').
	// +kubebuilder:default=none
	DataLoad string `json:"dataLoad,omitempty"`
}

type OfbizDBSpec struct {
	// Database name.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// The user which privileges to own ofbiz database
	// +kubebuilder:validation:Required
	User string `json:"user"`

	// The 'password' used for database user.
	// +kubebuilder:validation:Required
	Password PasswordSource `json:"password"`
}

// PostgresAdminSpec holds credentials for the operator to manage the database.
// If this section is provided, the operator will attempt to create the database and user.
type PostgresAdminSpec struct {
	// postgre super user
	// +kubebuilder:validation:Required
	User string `json:"user"`

	// The name of the secret containing the admin user's password.
	// The secret must have a key named 'password'.
	// +kubebuilder:validation:Required
	PasswordSecretName string `json:"passwordSecretName"`
}

// WebSpec defines web server configuration.
type WebSpec struct {
	SslSecretName string `json:"sslSecretName,omitempty"`
}

// AdminAccountSpec defines the initial admin credentials.
type AdminAccountSpec struct {
	Password PasswordSource `json:"password"`
}

// VolumeMounts for OFBiz
//      - ./ofbiz/docker-entrypoint-hooks:/docker-entrypoint-hooks:rw
//      - ./ofbiz/config:/ofbiz/config:rw
//      - ./ofbiz/lib-extra:/ofbiz/lib-extra:rw
//      - ./ofbiz/runtime:/ofbiz/runtime:rw
//      - ./ofbiz/export:/ofbiz/export:rw
//      - ./ofbiz/plugins:/ofbiz/plugins:rw

// StorageSpec defines the persistence configuration.
type StorageSpec struct {
	// Source for the OFBiz configuration files.
	// +optional
	Configuration *ConfigurationSource `json:"configuration,omitempty"`
	Persistence   PersistenceSpec      `json:"persistence,omitempty"`
}

// PersistenceSpec defines the PVC for stateful data.
type PersistenceSpec struct {
	Enabled          bool              `json:"enabled"`
	StorageClassName string            `json:"storageClassName,omitempty"`
	Size             resource.Quantity `json:"size"`
}

// OfbizStatus defines the observed state of Ofbiz
type OfbizStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	Nodes      []string           `json:"nodes,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type Ofbiz struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              OfbizSpec   `json:"spec,omitempty"`
	Status            OfbizStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

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
