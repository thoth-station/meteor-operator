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
	// Time to live after the resource was created. If empty default ttl will be enforced.
	//+kubebuilder:default=86400
	TTL int64 `json:"ttl"`
}

type MeteorImage struct {
	// ImageStream name. Empty if not yet created.
	//+optional
	ImageStreamName string `json:"name,omitempty"`
	// Url to a running deployment. Routable at least within the cluster. Empty if not yet scheduled.
	//+optional
	Url string `json:"url,omitempty"`
	// True if build completed successfully.
	//+optional
	Ready string `json:"ready,omitempty"`
}

// MeteorStatus defines the observed state of Meteor
type MeteorStatus struct {
	// Current condition of the Meteor.
	//+optional
	Phase string `json:"phase,omitempty"`
	// Current service state of Meteor.
	//+optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// A human readable message indicating details about why the Meteor is in this condition.
	//+optional
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the Meteor is in this state.
	// e.g. 'DiskPressure'
	//+optional
	Reason string `json:"reason,omitempty"`
	// JupyterBook deployment of Meteor. Empty if not created.
	//+optional
	JupyterBook MeteorImage `json:"jupyterBook,omitempty"`
	// JupyterHub image of Meteor. Empty if not created.
	//+optional
	JupyterHub MeteorImage `json:"jupyterHub,omitempty"`
	// Once created the expiration clock starts ticking.
	//+optional
	ExpireAt metav1.Time `json:"expireAt,omitempty"`
	// Most recent observed generation of Meteor. Sanity check.
	//+optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Url",type="string",JSONPath=".spec.url",description="Repository URL"
//+operator-sdk:csv:customresourcedefinitions:resources={{PipelineRun,tekton.dev},{Deployment,apps},{Service,v1},{Route,route.openshift.io},{ImageStream,image.openshift.io}}

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

// Add to scheme
func init() {
	SchemeBuilder.Register(&Meteor{}, &MeteorList{})
}

// Return true if TTL is reached
func (m *Meteor) IsTTLReached() bool {
	return m.GetCreationTimestamp().Add(time.Duration(m.Spec.TTL) * time.Second).Before(time.Now())
}

const (
	MeteorPipelineLabel   = "meteor.operate-first.cloud/pipeline"
	MeteorDeploymentLabel = "meteor.operate-first.cloud/deployment"
	MeteorLabel           = "meteor.operate-first.cloud/meteor"
	ODHJupyterHubLabel    = "opendatahub.io/notebook-image"
)

// Pre-populate labels for children resources
func (m *Meteor) SeedLabels() map[string]string {
	return map[string]string{MeteorLabel: string(m.GetUID())}
}

const (
	PhaseFailed  = "Failed"
	PhaseRunning = "Building"
	PhaseOk      = "Ready"
)

// Aggregate phase from conditions
func (m *Meteor) AggregatePhase() string {
	if len(m.Status.Conditions) == 0 {
		return PhaseRunning
	}

	for _, c := range m.Status.Conditions {
		if c.Status == "False" {
			return PhaseFailed
		}

		// Claim ready only if pipelineruns have completed
		if strings.HasPrefix(c.Type, "PipelineRun") {
			switch c.Reason {
			case "Succeeded", "Completed":
				continue
			}
			return PhaseRunning
		}

		if c.Reason != "Ready" {
			return PhaseRunning
		}
	}
	return PhaseOk
}
