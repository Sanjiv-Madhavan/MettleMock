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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PostgresDatabaseInstanceSpec defines the desired state of PostgresDatabaseInstance.
type PostgresDatabaseInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PostgresDatabaseInstance. Edit postgresdatabaseinstance_types.go to remove/update
	// +kubebuilder:validation:required:=true
	DatabaseName string `json:"db_name"`

	// +kubebuilder:validation:required:=true
	UserName string `json:"userName"`

	// +kubebuilder:default="0s"
	RotationInterval string `json:"rotation_interval,omitempty"`

	// +kubebuilder:default:=true
	DropOnDelete bool `json:"drop_on_delete,omitempty"`
}

// PostgresDatabaseInstanceStatus defines the observed state of PostgresDatabaseInstance.
type PostgresDatabaseInstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase      string `json:"phase,omitempty"`
	Message    string `json:"message,omitempty"`
	SecretName string `json:"secret,omitempty"`
}

// type PostgresDatabaseInstanceConditionType string

// type Conditions []PostgresDatabaseInstanceConditions

// type PostgresDatabaseInstanceConditions struct {
// 	// Type of condition.
// 	// +required
// 	Type PostgresDatabaseInstanceConditionType `json:"type" description:"type of status condition"`
// 	// Status of the condition, one of True, False, Unknown.
// 	// +required
// 	Status corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`

// 	// Reason for the condition's last transition
// 	// +optional
// 	Reason string `json:"reason,omitempty" description:"one-word CamelCase reason for the condition's last transition"`

// 	// A human readable message indicating details about the transition.
// 	// +optional
// 	Message string `json:"message,omitempty" description:"human-readable message indicating details about last transition"`

// 	// This flag indicates if the step failed.
// 	// +optional
// 	Failed bool `json:"failed,omitempty" description:"indicates if this step failed"`
// }

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PostgresDatabaseInstance is the Schema for the postgresdatabaseinstances API.
type PostgresDatabaseInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresDatabaseInstanceSpec   `json:"spec,omitempty"`
	Status PostgresDatabaseInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgresDatabaseInstanceList contains a list of PostgresDatabaseInstance.
type PostgresDatabaseInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresDatabaseInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgresDatabaseInstance{}, &PostgresDatabaseInstanceList{})
}
