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

	corev1 "k8s.io/api/core/v1"
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
						Status: corev1.ConditionTrue,
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
						Status: corev1.ConditionFalse,
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
						Status: corev1.ConditionUnknown,
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
		expectedOutput CNBiPhase
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
			expectedOutput: CNBiPhasePending,
		},

		"ok": {
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
							Type:   "PipelineRunPrepare",
							Status: corev1.ConditionTrue,
							Reason: "Succeeded",
						},
					},
				},
			},
			expectedOutput: CNBiPhaseSucceeded,
		},
		"failed": {
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
							Status: corev1.ConditionTrue,
							Reason: "ErrorCreatingPipelineRun",
						},
					},
				},
			},
			expectedOutput: CNBiPhaseFailed,
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
							Status: corev1.ConditionTrue,
							Reason: "ImportPipelineRunCreated",
						},
					},
				},
			},
			expectedOutput: CNBiPhaseImporting,
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
							Status: corev1.ConditionTrue,
							Reason: "ImportPipelineRunCreated",
						},
						{
							Type:   RequiredSecretMissing,
							Status: corev1.ConditionTrue,
							Reason: "ImapgePullSecretMissing",
						},
					},
				},
			},
			expectedOutput: CNBiPhaseImporting,
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
							Status: corev1.ConditionTrue,
							Reason: "ImportPipelineRunCreated",
						},
						{
							Type:   ValidatingImportedImage,
							Status: corev1.ConditionTrue,
							Reason: "ValidatingImportedImage",
						},
					},
				},
			},
			expectedOutput: CNBiPhaseValidating,
		},
	}

	for tcName, tc := range testCases {
		if output := tc.cnbi.AggregatePhase(); output != tc.expectedOutput {
			t.Errorf("%s Got %s while expecting %s", tcName, output, tc.expectedOutput)
		}
	}
}
