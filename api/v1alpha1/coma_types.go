/*
Copyright 2021.

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

// ComaSpec defines the desired state of Coma
type ComaSpec struct {
}

// ComaStatus defines the observed state of Coma
type ComaStatus struct {
	// Meteor owning this coma in a different namespace
	Owner NamespacedOwnerReference `json:"owner"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Coma is a complementary resource to Meteor in namespaces defined by Shower's externalServices property. This resource is generated automatically.
type Coma struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComaSpec   `json:"spec,omitempty"`
	Status ComaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ComaList contains a list of Coma
type ComaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Coma `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Coma{}, &ComaList{})
}
