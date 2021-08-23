package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Submit a Tekton PipelineRun from a collection
func (r *MeteorReconciler) ReconcilePipelineRun(name string, ctx *context.Context, req ctrl.Request) error {
	res := &pipelinev1beta1.PipelineRun{}
	resourceName := fmt.Sprintf("%s-%s", r.Meteor.GetName(), name)
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("pipelinerun", namespacedName)

	labels := r.Meteor.SeedLabels()
	labels[meteorv1alpha1.MeteorPipelineLabel] = name

	updateStatus := func(status metav1.ConditionStatus, reason, message string) {
		r.UpdateStatus(r.Meteor, "PipelineRun", name, status, reason, message)
	}

	statusIndex := func() int {
		for i, pr := range r.Meteor.Status.Pipelines {
			if pr.Name == name {
				return i
			}
		}
		result := meteorv1alpha1.PipelineResult{
			Name:  name,
			Ready: "False",
		}
		r.Meteor.Status.Pipelines = append(r.Meteor.Status.Pipelines, result)
		return len(r.Meteor.Status.Pipelines) - 1
	}()

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
								StringVal: r.Meteor.Spec.Url,
							},
						},
						{
							Name: "ref",
							Value: pipelinev1beta1.ArrayOrString{
								Type:      pipelinev1beta1.ParamTypeString,
								StringVal: r.Meteor.Spec.Ref,
							},
						},
					},
					Workspaces: []pipelinev1beta1.WorkspaceBinding{
						{
							Name: "data",
							VolumeClaimTemplate: &v1.PersistentVolumeClaim{
								Spec: v1.PersistentVolumeClaimSpec{
									AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
									Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											v1.ResourceStorage: resource.MustParse("500Mi"),
										},
									},
								},
							},
						},
					},
				},
			}
			controllerutil.SetControllerReference(r.Meteor, res, r.Scheme)

			if err := r.Create(*ctx, res); err != nil {
				logger.Error(err, "Unable to create PipelineRun")
				updateStatus(metav1.ConditionTrue, "CreateError", fmt.Sprintf("Unable to create pipelinerun. %s", err))
				return err
			}
			updateStatus(metav1.ConditionTrue, "BuildStated", "Tekton pipeline was submitted.")
			return nil
		}
		logger.Error(err, "Error fetching PipelineRun")

		updateStatus(metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if len(res.Status.Conditions) > 0 {
		if len(res.Status.Conditions) != 1 {
			logger.Error(nil, "Tekton reported multiple conditions")
		}
		condition := res.Status.Conditions[0]
		updateStatus(metav1.ConditionStatus(condition.Status), condition.Reason, condition.Message)
	}
	if res.Status.CompletionTime != nil && res.Status.Conditions[0].Reason == "Succeeded" {
		r.Meteor.Status.Pipelines[statusIndex].Ready = "True"
		if len(res.Status.PipelineResults) > 0 {
			r.Meteor.Status.Pipelines[statusIndex].Url = res.Status.PipelineResults[0].Value
		}
	}
	return nil
}
