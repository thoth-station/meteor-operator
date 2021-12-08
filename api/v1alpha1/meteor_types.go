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
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make" to regenerate code after modifying this file

// MeteorSpec defines the desired state of Meteor
type MeteorSpec struct {
	// Url points to the source repository.
	//+kubebuilder:validation:Pattern=`^https?:\/\/.+$`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Repository URL",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Url string `json:"url"`
	// Branch or tag or commit reference within the repository.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Branch Reference",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Ref string `json:"ref"`
	// Time to live after the resource was created.
	//+optional
	TTL int64 `json:"ttl,omitempty"`
	// List of pipelines to initiate for this meteor
	//+kubebuilder:default={jupyterhub,jupyterbook}
	Pipelines []string `json:"pipelines"`
}

type PipelineResult struct {
	Name string `json:"name"`
	// Url to a running deployment. Routable at least within the cluster. Empty if not yet scheduled.
	//+optional
	Url string `json:"url,omitempty"`
	// True if build completed successfully.
	//+optional
	Ready string `json:"ready,omitempty"`
}

type NamespacedOwnerReference struct {
	metav1.OwnerReference `json:",inline"`
	// Namespace of the resource
	Namespace string `json:"namespace"`
}

// MeteorStatus defines the observed state of Meteor
type MeteorStatus struct {
	// Current condition of the Meteor.
	//+optional
	Phase string `json:"phase,omitempty"`
	// Current service state of Meteor.
	//+optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// Stores results from pipelines. Empty if neither pipeline has completed.
	//+optional
	Pipelines []PipelineResult `json:"pipelines,omitempty"`
	// Once created the expiration clock starts ticking.
	//+optional
	ExpirationTimestamp metav1.Time `json:"expirationTimestamp,omitempty"`
	// Most recent observed generation of Meteor. Sanity check.
	//+optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// List of comas owned in different namespaces
	//+optional
	Comas []NamespacedOwnerReference `json:"comas,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"
//+kubebuilder:printcolumn:name="Url",type="string",JSONPath=".spec.url",description="Repository URL"
//+operator-sdk:csv:customresourcedefinitions:resources={{PipelineRun,v1beta1},{Deployment,v1},{Service,v1},{Route,v1},{ImageStream,v1}}

// Meteor resource represents a repository build. It defines which pipelines are executed and what is the livespan of the produced resources
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

// Add to scheme
func init() {
	SchemeBuilder.Register(&Meteor{}, &MeteorList{})
}

// Return true if TTL is reached
func (m *Meteor) IsTTLReached() bool {
	if m.Spec.TTL == 0 { // TTL not set
		return false
	}
	return m.GetExpirationTimestamp().Before(time.Now())
}

func (m *Meteor) GetRemainingTTL() float64 {
	return time.Until(m.GetExpirationTimestamp()).Seconds()
}

func (m *Meteor) GetExpirationTimestamp() time.Time {
	return m.GetCreationTimestamp().Add(time.Duration(m.Spec.TTL) * time.Second)
}

// Aggregate phase from conditions
func (m *Meteor) AggregatePhase() string {
	if len(m.Status.Conditions) == 0 {
		return PhaseBuilding
	}

	for _, c := range m.Status.Conditions {
		if c.Status == metav1.ConditionFalse {
			return PhaseFailed
		}

		// Claim ready only if pipelineruns have completed
		if strings.HasPrefix(c.Type, "PipelineRun") {
			switch c.Reason {
			case "Succeeded", "Completed":
				continue
			}
			return PhaseBuilding
		}

		if c.Reason != "Ready" {
			return PhaseBuilding
		}
	}
	return PhaseOk
}

func (m *Meteor) GetReference(isController bool) NamespacedOwnerReference {
	blockOwnerDeletion := true
	return NamespacedOwnerReference{
		OwnerReference: metav1.OwnerReference{
			APIVersion:         m.APIVersion,
			Kind:               m.Kind,
			Name:               m.GetName(),
			UID:                m.GetUID(),
			Controller:         &isController,
			BlockOwnerDeletion: &blockOwnerDeletion,
		},
		Namespace: m.GetNamespace(),
	}
}
