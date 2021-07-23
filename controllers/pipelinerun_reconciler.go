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
func (r *MeteorReconciler) ReconcilePipelineRun(pipelineName string, ctx *context.Context, req ctrl.Request, meteor *meteorv1alpha1.Meteor, status *meteorv1alpha1.MeteorImage) error {
	logger := log.FromContext(*ctx)
	pipelineRunName := fmt.Sprintf("meteor-%s-%s", pipelineName, meteor.GetUID())
	pipelineRun := &pipelinev1beta1.PipelineRun{}

	if err := r.Get(*ctx, types.NamespacedName{Name: pipelineRunName, Namespace: req.NamespacedName.Namespace}, pipelineRun); err != nil {
		if errors.IsNotFound(err) {
			pipelineRun = &pipelinev1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pipelineRunName,
					Namespace: req.NamespacedName.Namespace,
					Labels: map[string]string{
						MeteorPipelineLabel: pipelineName,
						MeteorLabel:         string(meteor.GetUID()),
					},
				},
				Spec: pipelinev1beta1.PipelineRunSpec{
					PipelineRef: &pipelinev1beta1.PipelineRef{
						Name: pipelineName,
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
			controllerutil.SetControllerReference(meteor, pipelineRun, r.Scheme)

			if err := r.Create(*ctx, pipelineRun); err != nil {
				logger.Error(err, fmt.Sprintf("Unable to instantiate %s pipelinerun", pipelineName))
				updatePipelineRunStatus(meteor, pipelineName, metav1.ConditionTrue, "CreateError", fmt.Sprintf("Unable to create pipelinerun. %s", err))
				return err
			}
			updatePipelineRunStatus(meteor, pipelineName, metav1.ConditionTrue, "BuildStated", "Tekton pipeline was submitted.")
			return nil
		}

		updatePipelineRunStatus(meteor, pipelineName, metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if len(pipelineRun.Status.Conditions) > 0 {
		if len(pipelineRun.Status.Conditions) != 1 {
			logger.Error(nil, "Tekton should not report more than one condition, using the first one only")
		}
		condition := pipelineRun.Status.Conditions[0]
		updatePipelineRunStatus(meteor, pipelineName, metav1.ConditionStatus(condition.Status), condition.Reason, condition.Message)
	}
	if pipelineRun.Status.CompletionTime != nil {
		// FIXME do only if succesfull
		status.Image = GetImageName(req.Namespace, pipelineName, meteor.GetUID())
	}
	return nil
}
