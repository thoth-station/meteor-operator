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

package cnbi

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	meteorv1alpha1 "github.com/thoth-station/meteor-operator/api/v1alpha1"
)

type ConditionsAware interface {
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
}

func AppendCondition(ctx *context.Context, reconcilerClient client.Client, object client.Object,
	typeName meteorv1alpha1.ConditionType, status metav1.ConditionStatus, reason string, message string) error {
	log := log.FromContext(*ctx)
	conditionsAware, conversionSuccessful := (object).(ConditionsAware)
	if conversionSuccessful {
		time := metav1.Time{Time: time.Now()}
		condition := metav1.Condition{Type: string(typeName), Status: status, Reason: reason, Message: message, LastTransitionTime: time}
		conditionsAware.SetConditions(append(conditionsAware.GetConditions(), condition))
		err := reconcilerClient.Status().Update(*ctx, object)
		if err != nil {
			errMessage := "Custom resource status update failed"
			log.Info(errMessage)
			return fmt.Errorf(errMessage)
		}

	} else {
		errMessage := "Status cannot be set, resource doesn't support conditions"
		log.Info(errMessage)
		return fmt.Errorf(errMessage)
	}
	return nil
}

func (reconciler *CustomNBImageReconciler) setConditionRequiredSecretMissing(ctx *context.Context,
	cnbi *meteorv1alpha1.CustomNBImage) error {

	if !reconciler.containsCondition(ctx, cnbi, string(meteorv1alpha1.RequiredSecretMissing)) {
		return AppendCondition(ctx, reconciler.Client, cnbi, meteorv1alpha1.RequiredSecretMissing, metav1.ConditionTrue,
			string(meteorv1alpha1.RequiredSecretMissing), string(meteorv1alpha1.RequiredSecretMissing))
	}
	return nil
}

func (reconciler *CustomNBImageReconciler) setConditionGenericPipelineError(ctx *context.Context, cnbi *meteorv1alpha1.CustomNBImage) {
	// updateStatus(metav1.ConditionFalse, "ErrorPipelineRun", fmt.Sprintf("Reconcile resulted in error. %s", err))
	panic("unimplemented")
}
func (reconciler *CustomNBImageReconciler) setConditionCreatePipelineRunError(ctx *context.Context, cnbi *meteorv1alpha1.CustomNBImage) error {
	if !reconciler.containsCondition(ctx, cnbi, string(meteorv1alpha1.ErrorPipelineRunCreate)) {
		return AppendCondition(ctx, reconciler.Client, cnbi, meteorv1alpha1.ErrorPipelineRunCreate, metav1.ConditionTrue,
			string(meteorv1alpha1.ErrorPipelineRunCreate), string(meteorv1alpha1.ErrorPipelineRunCreate))
	}
	return nil

}
func (reconciler *CustomNBImageReconciler) setConditionPipelineRunCreated(ctx *context.Context, cnbi *meteorv1alpha1.CustomNBImage) error {
	if !reconciler.containsCondition(ctx, cnbi, string(meteorv1alpha1.PipelineRunCreated)) {
		return AppendCondition(ctx, reconciler.Client, cnbi, meteorv1alpha1.PipelineRunCreated, metav1.ConditionTrue,
			string(meteorv1alpha1.PipelineRunCreated), string(meteorv1alpha1.PipelineRunCreated))
	}
	return nil

}

func (reconciler *CustomNBImageReconciler) containsCondition(ctx *context.Context,
	cnbi *meteorv1alpha1.CustomNBImage, reason string) bool {

	output := false
	for _, condition := range cnbi.Status.Conditions {
		if condition.Reason == reason {
			output = true
		}
	}
	return output
}
