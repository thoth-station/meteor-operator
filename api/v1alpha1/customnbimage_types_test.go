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
		status         CustomNBImageStatus
		expectedOutput bool
	}{
		"readySucceeded": {
			status: CustomNBImageStatus{
				Conditions: []metav1.Condition{
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
			status: CustomNBImageStatus{
				Conditions: []metav1.Condition{
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
			status: CustomNBImageStatus{
				Conditions: []metav1.Condition{
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
