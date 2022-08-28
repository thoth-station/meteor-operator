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

// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
const (
	CNBiPhasePending            = "Pending"
	CNBiPhaseFailed             = "Failed"
	CNBiPhasePreparing          = "Preparing"
	CNBiPhaseCreatingRepository = "CreatingRepository"
	CNBiPhaseResolving          = "Resolving"
	CNBiPhaseRunning            = "Running"
	CNBiPhaseBuilding           = "Building"
	CNBiPhageImporting          = "Importing"
	CNBiPhaseOk                 = "Ready"
)

// Strategies we support for a CNBi
const (
	CNBiStrategyImageImport         = "import"
	CNBiStrategyBuildUsingPython    = "build"
	CNBiStrategyBuildUsingBaseImage = "baseImage"
)

// CustomNBImageStrategy is the strategy super-set of configurations for all strategies.
type CustomNBImageStrategy struct {
	// Type is the strategy
	Type string `json:"type,omitempty"`
	// From is the reference to the source image, used for import strategy
	From string `json:"from,omitempty"`
}

// CustomNBImageRuntimeSpec defines a Runtime Environment, aka 'the Python version used'
type CustomNBImageRuntimeSpec struct {
	// PythonVersion is the version of Python to use
	PythonVersion string `json:"pythonVersion,omitempty"`
	// OSName is the Name of the Operating System to use
	OSName string `json:"osName,omitempty"`
	// OSVersion is the Version of the Operating System to use
	OSVersion string `json:"osVersion,omitempty"`
	// BaseImage is an alternative the the three above fields
	BaseImage string `json:"baseImage,omitempty"`
}

// CustomNBImageDashboardInformation is the information needed to generate the annoatation used by Open Data Hub Dashboard.
type CustomNBImageDashboardInformation struct {
	// Name is the name of the CustomNBImage
	Name string `json:"name"`
	// Description is the description of the CustomNBImage
	Description string `json:"description,omitempty"`
	// Creator is the name of the user who created the CustomNBImage
	Creator string `json:"creator"`
}

// CustomNBImageSpec defines the desired state of CustomNBImage
type CustomNBImageSpec struct {
	// RuntimeEnvironment is the runtime environment to use for the Custom Notebook Image
	RuntimeEnvironment CustomNBImageRuntimeSpec `json:"runtimeEnvironment,omitempty"`
	// PackageVersion is a set of Packages including their Version Specifiers
	PackageVersion []string `json:"packageVersions,omitempty"`
	// DashboardInformation is the information needed to generate the annoatation used by Open Data Hub Dashboard.
	DashboardInformation CustomNBImageDashboardInformation `json:"dashboardInformation"`
	// StrategyConfig is the configuration for the strategy, if no strategy is specified, we assume "prepare"
	Strategy CustomNBImageStrategy `json:"strategy,omitempty"`
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
	// Stores results from pipelines. Empty if neither pipeline has completed.
	//+optional
	Pipelines []PipelineResult `json:"pipelines,omitempty"`
}

// Aggregate phase from conditions
func (cnbi *CustomNBImage) AggregatePhase() string {
	if len(cnbi.Status.Conditions) == 0 {
		return CNBiPhasePending
	}

	for _, c := range cnbi.Status.Conditions {
		if c.Status == metav1.ConditionFalse {
			return CNBiPhaseFailed
		}

		// Claim ready only if pipelineruns have completed
		if strings.HasPrefix(c.Type, "PipelineRun") {
			switch c.Reason {
			case "Succeeded", "Completed":
				continue
			}
			// TODO distinguish between preparing and building (depending on the pipelinerun name containing 'prepare' or 'build'?)
			if strings.HasPrefix(c.Type, "PipelineRunPrepare") {
				return CNBiPhasePreparing
			} else if strings.HasPrefix(c.Type, "PipelineRunImport") {
				return CNBiPhageImporting
			}
		}

		if c.Reason != "Ready" {
			return CNBiPhaseRunning
		}
	}
	return CNBiPhaseOk
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"

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

// IsReady returns true the Ready condition status is True
func (status CustomNBImageStatus) IsReady() bool {
	for _, condition := range status.Conditions {
		if condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}
