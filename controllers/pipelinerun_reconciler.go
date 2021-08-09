package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updatePipelineRunStatus(meteor *meteorv1alpha1.Meteor, name string, status metav1.ConditionStatus, reason, message string) {
	updateStatus(meteor, "PipelineRun", name, status, reason, message)
}

// Submit a Tekton PipelineRun from a collection
func (r *MeteorReconciler) ReconcilePipelineRun(name string, ctx *context.Context, req ctrl.Request, meteor *meteorv1alpha1.Meteor, status *meteorv1alpha1.MeteorImage) error {
	res := &pipelinev1beta1.PipelineRun{}
	resourceName := fmt.Sprintf("%s-%s", meteor.GetName(), name)
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("pipelinerun", namespacedName)

	labels := MeteorLabels(meteor)
	labels[MeteorPipelineLabel] = name
	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating PipelineRun")

			res = &pipelinev1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
					Labels:    labels,
				},
				Spec: pipelinev1beta1.PipelineRunSpec{
					PipelineRef: &pipelinev1beta1.PipelineRef{
						Name: name,
					},
					Params: []pipelinev1beta1.Param{
						{
							Name: "url",
							Value: pipelinev1beta1.ArrayOrString{
								Type:      pipelinev1beta1.ParamTypeString,
								StringVal: meteor.Spec.Url,
							},
						},
						{
							Name: "ref",
							Value: pipelinev1beta1.ArrayOrString{
								Type:      pipelinev1beta1.ParamTypeString,
								StringVal: meteor.Spec.Ref,
							},
						},
						{
							Name: "uid",
							Value: pipelinev1beta1.ArrayOrString{
								Type:      pipelinev1beta1.ParamTypeString,
								StringVal: string(meteor.GetUID()),
							},
						},
					},
				},
			}
			controllerutil.SetControllerReference(meteor, res, r.Scheme)

			if err := r.Create(*ctx, res); err != nil {
				logger.Error(err, "Unable to create PipelineRun")
				updatePipelineRunStatus(meteor, name, metav1.ConditionTrue, "CreateError", fmt.Sprintf("Unable to create pipelinerun. %s", err))
				return err
			}
			updatePipelineRunStatus(meteor, name, metav1.ConditionTrue, "BuildStated", "Tekton pipeline was submitted.")
			return nil
		}
		logger.Error(err, "Error fetching PipelineRun")

		updatePipelineRunStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if len(res.Status.Conditions) > 0 {
		if len(res.Status.Conditions) != 1 {
			logger.Error(nil, "Tekton reported multiple conditions")
		}
		condition := res.Status.Conditions[0]
		updatePipelineRunStatus(meteor, name, metav1.ConditionStatus(condition.Status), condition.Reason, condition.Message)
	}
	if res.Status.CompletionTime != nil && res.Status.Conditions[0].Reason == "Succeeded" {
		status.Ready = "True"
	}
	return nil
}
