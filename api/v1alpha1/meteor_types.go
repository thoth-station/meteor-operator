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
	TTL int `json:"ttl"`
}

type MeteorImage struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type MeteorCondition struct {
	// Type is the type of the condition.
	// Currently only JupyterHub, JupyterBook, Deployment.
	Type string `json:"type"`
	// Status is the status of the condition.
	// Can be Unknown, Ready, Progressing, Failed.
	Status string `json:"status"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// MeteorStatus defines the observed state of Meteor
type MeteorStatus struct {
	// Current condition of the Meteor.
	// +optional
	Phase string `json:"phase,omitempty"`
	// Current service state of Meteor.
	// +optional
	Conditions []MeteorCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// A human readable message indicating details about why the Meteor is in this condition.
	// +optional
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the Meteor is in this state.
	// e.g. 'DiskPressure'
	// +optional
	Reason string `json:"reason,omitempty"`
	// JupyterBook host for the Meteor. Routable at least within the cluster. Empty if not yet scheduled.
	// +optional
	JupyterBook string `json:"jupyterBook,omitempty"`
	// JupyterHub ImageStream name for the Meteor. Empty if not yet created.
	// +optional
	JupyterHub string `json:"jupyterHub,omitempty"`
	// Images built from the source for this Meteor. Empty if no image is built yet.
	// +optional
	Images []MeteorImage `json:"images,omitempty"`
	// Once created the expiration clock starts ticking.
	// +optional
	ExpireAt metav1.Time `json:"expireAt,omitempty"`
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

// Filter Meteor images by name
func (m *Meteor) FilterImages(name string) *MeteorImage {
	for _, mi := range m.Status.Images {
		if mi.Name == name {
			return &mi
		}
	}
	return nil
}

func (m *Meteor) FilterConditions(name string) *MeteorCondition {
	for _, mc := range m.Status.Conditions {
		if mc.Type == name {
			return &mc
		}
	}
	return nil
}

func (m *Meteor) ReadyToStartPipeline(name string) bool {
	return m.FilterImages(name) == nil && m.FilterConditions(name).Status == "Unknown"
}
