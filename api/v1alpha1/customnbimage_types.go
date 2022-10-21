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

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// ImagePullSecret is a secret that is used to pull images from a private registry
type ImagePullSecret struct {
	// Name of the secret to be used
	Name string `json:"name"`
}

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
	// ImagePullSecret is the name of the secret to use for pulling the base image
	// +optional
	ImagePullSecret ImagePullSecret `json:"imagePullSecret,omitempty"`
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

// CustomNotebookImageStatus defines the observed state of CustomNBImage
type CustomNotebookImageStatus struct {
	// Current condition of the Shower.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Phase",xDescriptors={"urn:alm:descriptor:io.kubernetes.phase'"}
	//+optional
	Phase CNBiPhase `json:"phase,omitempty"`
	// Current service state of Meteor.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	//+optional
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// Stores results from pipelines. Empty if neither pipeline has completed.
	//+optional
	Pipelines []PipelineResult `json:"pipelines,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=cnbi,categories=opendatahub
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"

// CustomNBImage is the Schema for the customnbimages API
type CustomNBImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomNBImageSpec         `json:"spec,omitempty"`
	Status CustomNotebookImageStatus `json:"status,omitempty"`
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
		/*
			if c.Status == corev1.ConditionFalse {
				return CNBiPhaseFailed
			}
		*/

		if c.Type == ErrorPipelineRunCreate && c.Status == corev1.ConditionTrue {
			return CNBiPhaseFailed
		}

		if c.Type == PipelineRunCreated && c.Status == corev1.ConditionTrue {
			if strings.HasPrefix(c.Reason, "ImportPipeline") {
				return CNBiPhaseImporting
			} else {
				return CNBiPhaseBuilding
			}
		}
	}

	return CNBiPhaseSucceeded
}

// IsReady returns true the Ready condition status is True
func (status CustomNotebookImageStatus) IsReady() bool {
	for _, condition := range status.Conditions {
		if condition.Status == corev1.ConditionTrue {
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

// hasValidImagePullSecret checks if the ImagePullSecret is valid, eg name is not empty
func (b *BuildTypeSpec) hasValidImagePullSecret() bool {
	if b.ImagePullSecret != (ImagePullSecret{}) {
		if b.ImagePullSecret.Name == "" {
			return false
		}
	}

	return true
}

// GetConditions returns the set of conditions for this object.
func (c *CustomNBImage) GetConditions() Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (c *CustomNBImage) SetConditions(conditions Conditions) {
	c.Status.Conditions = conditions
}

func (c ConditionType) hasPrefix(prefix string) bool {
	return strings.HasPrefix(string(c), prefix)
}
