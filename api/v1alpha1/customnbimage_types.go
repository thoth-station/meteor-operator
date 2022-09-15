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

// CNBiPhase describes the phase of the CustomNBImage
// +kubebuilder:validation:Enum=Pending;Failed;Preparing;CreatingRepository;Resolving;Running;Building;Importing;Validating;Ready
type CNBiPhase string

const (
	CNBiPhasePending            CNBiPhase = "Pending"
	CNBiPhaseFailed             CNBiPhase = "Failed"
	CNBiPhasePreparing          CNBiPhase = "Preparing"
	CNBiPhaseCreatingRepository CNBiPhase = "CreatingRepository"
	CNBiPhaseResolving          CNBiPhase = "Resolving"
	CNBiPhaseRunning            CNBiPhase = "Running"
	CNBiPhaseBuilding           CNBiPhase = "Building"
	CNBiPhaseImporting          CNBiPhase = "Importing"
	CNBiPhaseValidating         CNBiPhase = "Validating"
	CNBiPhaseOk                 CNBiPhase = "Ready"
)

// BuildType describes how to build a custom notebook image.
// Only one of the following build types may be specified.
// +kubebuilder:validation:MinLength:1
// +kubebuilder:validation:Enum=ImageImport;PackageList;GitRepository
type BuildType string

const (
	// ImportImage will simply import the image from the given URL
	ImportImage BuildType = "ImageImport"

	// PackageList will build a custom image using a specific List of Python Packages
	// if no RuntimeEnvironment is specified, a baseImage must be specified for the build
	PackageList BuildType = "PackageList"

	// BuildGitRepository will builds a custom image from a git repository
	GitRepository BuildType = "GitRepository"
)

// CNBi Annotations is a list of annotations that are added to the custom notebook image
const (
	CNBiNameAnnotationKey        = "opendatahub.io/notebook-image-name"
	CNBiDescriptionAnnotationKey = "opendatahub.io/notebook-image-desc"
	CNBiCreatorAnnotationKey     = "opendatahub.io/notebook-image-creator"
)

// BuildTypeSpec is the strategy super-set of configurations for all strategies.
type BuildTypeSpec struct {
	// BuildType is the strategy
	// +required
	// +kubebuilder:Required
	BuildType BuildType `json:"buildType"`
	// FromImage is the reference to the source image, used for import strategy
	// +optional
	FromImage string `json:"fromImage,omitempty"`
	// BaseImage is the reference to the base image, used for building
	// +optional
	BaseImage string `json:"baseImage,omitempty"`
	// Repository is the URL of the git repository, used for building
	// +optional
	Repository string `json:"repository,omitempty"`
	// GitRef is the git reference within the Repository to use for building (e.g. "main")
	// +optional
	GitRef string `json:"gitRef,omitempty"`
}

// CustomNBImageRuntimeSpec defines a Runtime Environment, aka 'the Python version used'
type CustomNBImageRuntimeSpec struct {
	// PythonVersion is the version of Python to use
	// +optional
	PythonVersion string `json:"pythonVersion,omitempty"`
	// OSName is the Name of the Operating System to use
	// +optional
	OSName string `json:"osName,omitempty"`
	// OSVersion is the Version of the Operating System to use
	// +optional
	OSVersion string `json:"osVersion,omitempty"`
}

// CustomNBImageSpec defines the desired state of CustomNBImage
type CustomNBImageSpec struct {
	// RuntimeEnvironment is the runtime environment to use for the Custom Notebook Image
	// +optional
	RuntimeEnvironment CustomNBImageRuntimeSpec `json:"runtimeEnvironment,omitempty"`
	// PackageVersions is a set of Packages including their Version Specifiers
	// +optional
	PackageVersions []string `json:"packageVersions,omitempty"`
	// BuildType is the configuration for the build
	// +required
	// +kubebuilder:Required
	BuildTypeSpec `json:",inline"`
}

// CustomNBImageStatus defines the observed state of CustomNBImage
type CustomNBImageStatus struct {
	// Current condition of the Shower.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Phase",xDescriptors={"urn:alm:descriptor:io.kubernetes.phase'"}
	//+optional
	Phase CNBiPhase `json:"phase,omitempty"`
	// Current service state of Meteor.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	//+optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// Stores results from pipelines. Empty if neither pipeline has completed.
	//+optional
	Pipelines []PipelineResult `json:"pipelines,omitempty"`
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

// Aggregate phase from conditions
func (cnbi *CustomNBImage) AggregatePhase() CNBiPhase {
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
				return CNBiPhaseImporting
			}
		}

		if c.Reason != "Ready" {
			return CNBiPhaseRunning
		}
	}
	return CNBiPhaseOk
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

func init() {
	SchemeBuilder.Register(&CustomNBImage{}, &CustomNBImageList{})
}

// isValid checks if the RuntimeEnvironment is valid
// This function has been completely generated by GitHub Copilot based on the comment above 🪩
func (r *CustomNBImageRuntimeSpec) isValid() bool {
	if r.PythonVersion == "" {
		return false
	}
	if r.OSName == "" {
		return false
	}
	if r.OSVersion == "" {
		return false
	}

	return true
}
