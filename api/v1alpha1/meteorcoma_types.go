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

// MeteorComaSpec defines the desired state of MeteorComa
type MeteorComaSpec struct {
}

// MeteorComaStatus defines the observed state of MeteorComa
type MeteorComaStatus struct {
	// Meteor owning this coma in a different namespace
	Owner NamespacedOwnerReference `json:"owner"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MeteorComa is the Schema for the meteorcomas API
type MeteorComa struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeteorComaSpec   `json:"spec,omitempty"`
	Status MeteorComaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MeteorComaList contains a list of MeteorComa
type MeteorComaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeteorComa `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MeteorComa{}, &MeteorComaList{})
}
