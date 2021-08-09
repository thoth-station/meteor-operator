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

// Important: Run "make" to regenerate code after modifying this file

// MeteorSpec defines the desired state of Meteor
type MeteorSpec struct {
	// Url points to the source repository.
	Url string `json:"url"`
	// Branch or tag or commit reference within the repository.
	Ref string `json:"ref"`
	// Time to live after the resource was created. If empty default ttl will be enforced.
	TTL int64 `json:"ttl"`
}

type MeteorImage struct {
	// ImageStream name. Empty if not yet created.
	// +optional
	ImageStreamName string `json:"name,omitempty"`
	// Url to a running deployment. Routable at least within the cluster. Empty if not yet scheduled.
	// +optional
	Url string `json:"url,omitempty"`
	// True if build completed successfully.
	// +optional
	Ready string `json:"ready,omitempty"`
}

// MeteorStatus defines the observed state of Meteor
type MeteorStatus struct {
	// Current condition of the Meteor.
	// +optional
	Phase string `json:"phase,omitempty"`
	// Current service state of Meteor.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// A human readable message indicating details about why the Meteor is in this condition.
	// +optional
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the Meteor is in this state.
	// e.g. 'DiskPressure'
	// +optional
	Reason string `json:"reason,omitempty"`
	// JupyterBook deployment of Meteor. Empty if not created.
	// +optional
	JupyterBook MeteorImage `json:"jupyterBook,omitempty"`
	// JupyterHub image of Meteor. Empty if not created.
	// +optional
	JupyterHub MeteorImage `json:"jupyterHub,omitempty"`
	// Once created the expiration clock starts ticking.
	// +optional
	ExpireAt metav1.Time `json:"expireAt,omitempty"`
	// Most recent observed generation of Meteor. Sanity check.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Meteor is the Schema for the meteors API
type Meteor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeteorSpec   `json:"spec,omitempty"`
	Status MeteorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MeteorList contains a list of Meteor
type MeteorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Meteor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Meteor{}, &MeteorList{})
}

func (m *Meteor) FilterConditions(name string) *metav1.Condition {
	for _, mc := range m.Status.Conditions {
		if mc.Type == name {
			return &mc
		}
	}
	return nil
}
