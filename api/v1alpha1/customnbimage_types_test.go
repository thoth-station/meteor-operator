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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionReady    = "Ready"
	ConditionNotReady = "NotReady"
)

// TestIsReady tests IsReady condition status function
func TestIsReady(t *testing.T) {
	testCases := map[string]struct {
		status         CustomNotebookImageStatus
		expectedOutput bool
	}{
		"readySucceeded": {
			status: CustomNotebookImageStatus{
				Conditions: []Condition{
					{
						Type:   "PipelineRunPrepare",
						Status: metav1.ConditionTrue,
						Reason: "Succeeded",
					},
				},
			},
			expectedOutput: true,
		},
		"notReadyCouldntGetPipeline": {
			status: CustomNotebookImageStatus{
				Conditions: []Condition{
					{
						Type:   "PipelineRunPrepare",
						Status: metav1.ConditionFalse,
						Reason: "CouldntGetPipeline",
					},
				},
			},
			expectedOutput: false,
		},
		"notReadyRunning": {
			status: CustomNotebookImageStatus{
				Conditions: []Condition{
					{
						Type:   "PipelineRunPrepare",
						Status: metav1.ConditionUnknown,
						Reason: "Running",
					},
				},
			},
			expectedOutput: false,
		},
	}

	for tcName, tc := range testCases {
		if output := tc.status.IsReady(); output != tc.expectedOutput {
			t.Errorf("%s Got %t while expecting %t", tcName, output, tc.expectedOutput)
		}
	}

}
func TestIsValid(t *testing.T) {
	testCases := map[string]struct {
		runtime        CustomNBImageRuntimeSpec
		expectedOutput bool
	}{
		"validRuntime": {
			runtime: CustomNBImageRuntimeSpec{
				PythonVersion: "3.9",
				OSName:        "ubi",
				OSVersion:     "9",
			},
			expectedOutput: true,
		},
		"invalidRuntime": {
			runtime: CustomNBImageRuntimeSpec{
				PythonVersion: "",
				OSName:        "",
				OSVersion:     "",
			},
			expectedOutput: false,
		},
	}

	for tcName, tc := range testCases {
		if output := tc.runtime.isValid(); output != tc.expectedOutput {
			t.Errorf("%s Got %t while expecting %t", tcName, output, tc.expectedOutput)
		}
	}

}

func TestHasValidImagePullSecretAName(t *testing.T) {
	testCases := map[string]struct {
		spec           BuildTypeSpec
		expectedOutput bool
	}{
		"hasImagePullSecretName": {
			spec: BuildTypeSpec{
				BuildType:       ImportImage,
				ImagePullSecret: ImagePullSecret{Name: "test"},
			},
			expectedOutput: true,
		},
		"noImagePullSecret": {
			spec: BuildTypeSpec{
				BuildType: ImportImage,
			},
			expectedOutput: true,
		},
	}

	for tcName, tc := range testCases {
		if output := tc.spec.hasValidImagePullSecret(); output != tc.expectedOutput {
			t.Errorf("%s Got %t while expecting %t", tcName, output, tc.expectedOutput)
		}
	}

}

// TestAggregatePhase tests if condition are aggregated into the correct phase
func TestAggregatePhase(t *testing.T) {
	testCases := map[string]struct {
		cnbi           CustomNBImage
		expectedOutput Phase
	}{
		"pending": {
			cnbi: CustomNBImage{
				Spec: CustomNBImageSpec{
					PackageVersions: []string{},
					BuildTypeSpec: BuildTypeSpec{
						BuildType: ImportImage,
						FromImage: "quay.io/thoth-station/s2i-minimal-py38-notebook:v0.2.2",
					},
				},
				Status: CustomNotebookImageStatus{
					Conditions: []Condition{},
				},
			},
			expectedOutput: PhasePending,
		},

		"pipeline-created": {
			cnbi: CustomNBImage{
				Spec: CustomNBImageSpec{
					PackageVersions: []string{},
					BuildTypeSpec: BuildTypeSpec{
						BuildType: ImportImage,
						FromImage: "quay.io/thoth-station/s2i-minimal-py38-notebook:v0.2.2",
					},
				},
				Status: CustomNotebookImageStatus{
					Conditions: []Condition{
						{
							Type:   PipelineRunCreated,
							Status: metav1.ConditionTrue,
							Reason: "Succeeded",
						},
					},
				},
			},
			expectedOutput: PhaseRunning,
		},
		"pipeline-create-failed": {
			cnbi: CustomNBImage{
				Spec: CustomNBImageSpec{
					PackageVersions: []string{},
					BuildTypeSpec: BuildTypeSpec{
						BuildType: ImportImage,
						FromImage: "quay.io/thoth-station/s2i-minimal-py38-notebook:v0.2.2",
					},
				},
				Status: CustomNotebookImageStatus{
					Conditions: []Condition{
						{
							Type:   ErrorPipelineRunCreate,
							Status: metav1.ConditionTrue,
							Reason: "ErrorCreatingPipelineRun",
						},
					},
				},
			},
			expectedOutput: PhaseFailed,
		},
		"importing": {
			cnbi: CustomNBImage{
				Spec: CustomNBImageSpec{
					PackageVersions: []string{},
					BuildTypeSpec: BuildTypeSpec{
						BuildType: ImportImage,
						FromImage: "quay.io/thoth-station/s2i-minimal-py38-notebook:v0.2.2",
					},
				},
				Status: CustomNotebookImageStatus{
					Conditions: []Condition{
						{
							Type:   PipelineRunCreated,
							Status: metav1.ConditionTrue,
							Reason: "ImportPipelineRunCreated",
						},
						{
							Type:   ImageImportReady,
							Status: metav1.ConditionFalse,
							Reason: "ImageImportNotReady",
						},
					},
				},
			},
			expectedOutput: PhaseRunning,
		},
		"importing_missing_secret": {
			cnbi: CustomNBImage{
				Spec: CustomNBImageSpec{
					PackageVersions: []string{},
					BuildTypeSpec: BuildTypeSpec{
						BuildType: ImportImage,
						FromImage: "quay.io/thoth-station/s2i-minimal-py38-notebook:v0.2.2",
					},
				},
				Status: CustomNotebookImageStatus{
					Conditions: []Condition{
						{
							Type:   PipelineRunCreated,
							Status: metav1.ConditionTrue,
							Reason: "ImportPipelineRunCreated",
						},
						{
							Type:   RequiredSecretMissing,
							Status: metav1.ConditionTrue,
							Reason: "ImapgePullSecretMissing",
						},
					},
				},
			},
			expectedOutput: PhaseRunning,
		},
		"validating": {
			cnbi: CustomNBImage{
				Spec: CustomNBImageSpec{
					PackageVersions: []string{},
					BuildTypeSpec: BuildTypeSpec{
						BuildType: ImportImage,
						FromImage: "quay.io/thoth-station/s2i-minimal-py38-notebook:v0.2.2",
					},
				},
				Status: CustomNotebookImageStatus{
					Conditions: []Condition{
						{
							Type:   PipelineRunCreated,
							Status: metav1.ConditionTrue,
							Reason: "ImportPipelineRunCreated",
						},
						{
							Type:   ValidatingImportedImage,
							Status: metav1.ConditionTrue,
							Reason: "ValidatingImportedImage",
						},
					},
				},
			},
			expectedOutput: PhaseRunning,
		},
		"import-successful": {
			cnbi: CustomNBImage{
				Spec: CustomNBImageSpec{
					PackageVersions: []string{},
					BuildTypeSpec: BuildTypeSpec{
						BuildType: ImportImage,
						FromImage: "quay.io/thoth-station/s2i-minimal-py38-notebook:v0.2.2",
					},
				},
				Status: CustomNotebookImageStatus{
					Conditions: []Condition{
						{
							Type:   ImageImportReady,
							Status: metav1.ConditionTrue,
							Reason: "ImageImportReady",
						},
						{
							Type:   PipelineRunCompleted,
							Status: metav1.ConditionTrue,
							Reason: "PipelineRunCompleted",
						},
					},
				},
			},
			expectedOutput: PhaseSucceeded,
		},
		"import-failed": {
			cnbi: CustomNBImage{
				Spec: CustomNBImageSpec{
					PackageVersions: []string{},
					BuildTypeSpec: BuildTypeSpec{
						BuildType: ImportImage,
						FromImage: "quay.io/thoth-station/s2i-minimal-py38-notebook:v0.2.2",
					},
				},
				Status: CustomNotebookImageStatus{
					Conditions: []Condition{
						/*
						  - lastTransitionTime: "2022-10-25T12:55:46Z"
						    message: import PipelineRun created successfully
						    reason: PipelineRunCreated
						    status: "True"
						    type: PipelineRunCreated
						  - lastTransitionTime: "2022-10-25T12:56:34Z"
						    message: Import failed
						    reason: ImageImportNotReady
						    status: "False"
						    type: ImageImportReady

						*/
						{
							Type:   ImageImportReady,
							Status: metav1.ConditionFalse,
							Reason: "ImageImportNotReady",
						},
						{
							Type:   PipelineRunCompleted,
							Status: metav1.ConditionTrue,
							Reason: "PipelineRunCompleted",
						},
					},
				},
			},
			expectedOutput: PhaseFailed,
		},
	}

	for tcName, tc := range testCases {
		if output := tc.cnbi.AggregatePhase(); output != tc.expectedOutput {
			t.Errorf("%s Got %s while expecting %s", tcName, output, tc.expectedOutput)
		}
	}
}
