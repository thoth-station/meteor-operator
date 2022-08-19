/*
Copyright 2021, 2022 The Meteor Authors.

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

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomNBImageSpec defines the desired state of CustomNBImage
type CustomNBImageSpec struct {
	// Foo is an example field of CustomNBImage. Edit customnbimage_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// CustomNBImageStatus defines the observed state of CustomNBImage
type CustomNBImageStatus struct {
	// Current condition of the Shower.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Phase",xDescriptors={"urn:alm:descriptor:io.kubernetes.phase'"}
	//+optional
	Phase string `json:"phase,omitempty"`
	// Current service state of Meteor.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	//+optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// Aggregate phase from conditions
func (cnbi *CustomNBImage) AggregatePhase() string {
	if len(cnbi.Status.Conditions) == 0 {
		return PhasePending
	}

	//	for _, c := range cnbi.Status.Conditions {
	//	}

	return PhaseOk
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CustomNBImage is the Schema for the customnbimages API
type CustomNBImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomNBImageSpec   `json:"spec,omitempty"`
	Status CustomNBImageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CustomNBImageList contains a list of CustomNBImage
type CustomNBImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomNBImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomNBImage{}, &CustomNBImageList{})
}
