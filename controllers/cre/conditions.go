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

package cre

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	meteorv1alpha1 "github.com/thoth-station/meteor-operator/api/v1alpha1"
)

func setCondition(cnbiStatus *meteorv1alpha1.CustomRuntimeEnvironmentStatus, ctype meteorv1alpha1.ConditionType, status metav1.ConditionStatus, reason, message string) {
	var c *meteorv1alpha1.Condition
	for i := range cnbiStatus.Conditions {
		if cnbiStatus.Conditions[i].Type == ctype {
			c = &cnbiStatus.Conditions[i]
		}
	}
	if c == nil {
		addCondition(cnbiStatus, ctype, status, reason, message)
	} else {
		// check if the condition has changed
		if c.Status == status && c.Reason == reason && c.Message == message {
			return
		}
		// it has not changed, so update the condition
		c.LastTransitionTime = metav1.Now()
		c.Status = status
		c.Reason = reason
		c.Message = message
	}
}

func addCondition(cnbiStatus *meteorv1alpha1.CustomRuntimeEnvironmentStatus, ctype meteorv1alpha1.ConditionType, status metav1.ConditionStatus, reason, message string) {
	now := metav1.Now()
	c := meteorv1alpha1.Condition{
		Type:               ctype,
		LastTransitionTime: now,
		Status:             status,
		Reason:             reason,
		Message:            message,
	}
	cnbiStatus.Conditions = append(cnbiStatus.Conditions, c)
}

func removeCondition(cnbiStatus *meteorv1alpha1.CustomRuntimeEnvironmentStatus, ctype meteorv1alpha1.ConditionType) {
	var newConditions []meteorv1alpha1.Condition
	for _, c := range cnbiStatus.Conditions {
		if c.Type != ctype {
			newConditions = append(newConditions, c)
		}
	}
	cnbiStatus.Conditions = newConditions
}
