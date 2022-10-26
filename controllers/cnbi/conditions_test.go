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

package cnbi

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	meteorv1alpha1 "github.com/thoth-station/meteor-operator/api/v1alpha1"
)

func TestAddCondition(t *testing.T) {
	testCases := map[string]struct {
		status       meteorv1alpha1.CustomNotebookImageStatus
		expectedSize int
	}{
		"zero": {
			status: meteorv1alpha1.CustomNotebookImageStatus{
				Conditions: []meteorv1alpha1.Condition{},
			},
			expectedSize: 1,
		},
		"one": {
			status: meteorv1alpha1.CustomNotebookImageStatus{
				Conditions: []meteorv1alpha1.Condition{
					{
						Type:    meteorv1alpha1.ConditionType("test"),
						Status:  metav1.ConditionTrue,
						Reason:  "test",
						Message: "test",
					},
				},
			},
			expectedSize: 2,
		},
	}

	for tcName, tc := range testCases {
		addCondition(&tc.status, meteorv1alpha1.ConditionType("test2"), metav1.ConditionFalse, "test2", "test2")
		if len(tc.status.Conditions) != tc.expectedSize {
			t.Errorf("%s Condition, added one... expecting %d", tcName, tc.expectedSize)
		}
	}

}

func TestSetCondition(t *testing.T) {
	testCases := map[string]struct {
		status       meteorv1alpha1.CustomNotebookImageStatus
		expectedSize int
	}{
		"zero": {
			status: meteorv1alpha1.CustomNotebookImageStatus{
				Conditions: []meteorv1alpha1.Condition{},
			},
			expectedSize: 1,
		},
		"one": {
			status: meteorv1alpha1.CustomNotebookImageStatus{
				Conditions: []meteorv1alpha1.Condition{
					{
						Type:    meteorv1alpha1.ConditionType("test"),
						Status:  metav1.ConditionTrue,
						Reason:  "test",
						Message: "test",
					},
				},
			},
			expectedSize: 2,
		},
		"two": {
			status: meteorv1alpha1.CustomNotebookImageStatus{
				Conditions: []meteorv1alpha1.Condition{
					{
						Type:    meteorv1alpha1.ConditionType("test2"),
						Status:  metav1.ConditionFalse,
						Reason:  "test2",
						Message: "test2",
					},
				},
			},
			expectedSize: 1,
		},
		"eine": { // if we got one on True, we reset it to False
			status: meteorv1alpha1.CustomNotebookImageStatus{
				Conditions: []meteorv1alpha1.Condition{
					{
						Type:    meteorv1alpha1.ConditionType("test2"),
						Status:  metav1.ConditionTrue,
						Reason:  "test2",
						Message: "test2",
					},
				},
			},
			expectedSize: 1,
		},
	}

	for tcName, tc := range testCases {
		setCondition(&tc.status, meteorv1alpha1.ConditionType("test2"), metav1.ConditionFalse, "test2", "test2")
		if len(tc.status.Conditions) != tc.expectedSize {
			t.Errorf("%s Condition, set one... expecting %d", tcName, tc.expectedSize)
		}
	}

}

func TestRemoveCondition(t *testing.T) {
	testCases := map[string]struct {
		status       meteorv1alpha1.CustomNotebookImageStatus
		expectedSize int
	}{
		"zero": {
			status: meteorv1alpha1.CustomNotebookImageStatus{
				Conditions: []meteorv1alpha1.Condition{},
			},
			expectedSize: 0,
		},
		"one-to-stay": {
			status: meteorv1alpha1.CustomNotebookImageStatus{
				Conditions: []meteorv1alpha1.Condition{
					{
						Type:    meteorv1alpha1.ConditionType("test"),
						Status:  metav1.ConditionTrue,
						Reason:  "test",
						Message: "test",
					},
				},
			},
			expectedSize: 1,
		},
		"one-to-be-zero": {
			status: meteorv1alpha1.CustomNotebookImageStatus{
				Conditions: []meteorv1alpha1.Condition{
					{
						Type:    meteorv1alpha1.ConditionType("test2"),
						Status:  metav1.ConditionTrue,
						Reason:  "test2",
						Message: "test2",
					},
				},
			},
			expectedSize: 0,
		},
	}

	for tcName, tc := range testCases {
		removeCondition(&tc.status, meteorv1alpha1.ConditionType("test2"))
		if len(tc.status.Conditions) != tc.expectedSize {
			t.Errorf("%s Condition, removed one... expecting %d", tcName, tc.expectedSize)
		}
	}

}
