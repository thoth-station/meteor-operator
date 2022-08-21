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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/aicoe/meteor-operator/api/v1alpha1"
	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
)

// CustomNBImageReconciler reconciles a CustomNBImage object
type CustomNBImageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	CNBi   *meteorv1alpha1.CustomNBImage
}

//+kubebuilder:rbac:groups=meteor.zone,resources=customnbimages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=meteor.zone,resources=customnbimages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=meteor.zone,resources=customnbimages/finalizers,verbs=update
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CustomNBImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	r.CNBi = &meteorv1alpha1.CustomNBImage{}

	if err := r.Get(ctx, req.NamespacedName, r.CNBi); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Unable to fetch reconciled resource")
		return ctrl.Result{Requeue: true}, err
	}

	if !r.CNBi.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Resource being delete, skipping further reconcile.")
		return ctrl.Result{}, nil
	}

	r.CNBi.Status.Phase = r.CNBi.AggregatePhase()
	logger.Info("Reconciling CustomNBImage", "phase", r.CNBi.Status.Phase)

	// TODO your logic here

	if err := r.ReconcilePipelineRun("cnbi-create-repo", &ctx, req); err != nil {
		return r.UpdateStatusNow(ctx, err)
	}

	return r.UpdateStatusNow(ctx, nil)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomNBImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&meteorv1alpha1.CustomNBImage{}).
		Complete(r)
}

// Force object status update. Returns a reconcile result
func (r *CustomNBImageReconciler) UpdateStatusNow(ctx context.Context, originalErr error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	if err := r.Status().Update(ctx, r.CNBi); err != nil {
		logger.WithValues("reason", err.Error()).Info("Unable to update status, retrying")
		return ctrl.Result{Requeue: true}, nil
	}
	if originalErr != nil {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, originalErr
	} else {
		return ctrl.Result{}, nil
	}
}

// Set status condition helper
func (r *CustomNBImageReconciler) SetCondition(kind, name string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&r.CNBi.Status.Conditions, metav1.Condition{
		Type:    kind + strings.Title(name),
		Status:  status,
		Reason:  reason,
		Message: message,
	})
}

// ReconcilePipelineRun will reconcile the pipeline run for the CustomNBImage.
func (r *CustomNBImageReconciler) ReconcilePipelineRun(name string, ctx *context.Context, req ctrl.Request) error {
	pipelineRun := &pipelinev1beta1.PipelineRun{}
	resourceName := fmt.Sprintf("%s-%s", r.CNBi.GetName(), name)
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("pipelinerun", namespacedName)

	updateStatus := func(status metav1.ConditionStatus, reason, message string) {
		r.SetCondition("PipelineRun", name, status, reason, message)
	}

	statusIndex := func() int {
		for i, pr := range r.CNBi.Status.Pipelines {
			if pr.Name == name {
				return i
			}
		}
		result := v1alpha1.PipelineResult{
			Name:            name,
			Ready:           "False",
			PipelineRunName: resourceName,
		}
		r.CNBi.Status.Pipelines = append(r.CNBi.Status.Pipelines, result)
		return len(r.CNBi.Status.Pipelines) - 1
	}()

	if err := r.Get(*ctx, namespacedName, pipelineRun); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating PipelineRun")

			pipelineRun = &pipelinev1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
				},
				Spec: pipelinev1beta1.PipelineRunSpec{
					PipelineRef: &pipelinev1beta1.PipelineRef{
						Name: name,
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
						{
							Name: "sslcertdir",
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "openshift-service-ca.crt",
								},
								Items: []v1.KeyToPath{{
									Key:  "service-ca.crt",
									Path: "ca.crt",
								}},
								DefaultMode: pointer.Int32(420),
							},
						},
					},
				},
			}
			controllerutil.SetControllerReference(r.CNBi, pipelineRun, r.Scheme)

			if err := r.Create(*ctx, pipelineRun); err != nil {
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

	if len(pipelineRun.Status.Conditions) > 0 {
		if len(pipelineRun.Status.Conditions) != 1 {
			logger.Error(nil, "Tekton reported multiple conditions")
		}
		condition := pipelineRun.Status.Conditions[0]
		updateStatus(metav1.ConditionStatus(condition.Status), condition.Reason, condition.Message)
	}

	if pipelineRun.Status.CompletionTime != nil {
		if pipelineRun.Status.Conditions[0].Reason == "Succeeded" {
			r.CNBi.Status.Pipelines[statusIndex].Ready = "True"
			if len(pipelineRun.Status.PipelineResults) > 0 {
				if pipelineRun.Status.PipelineResults[0].Value.Type == "string" {
					r.CNBi.Status.Pipelines[statusIndex].Url = pipelineRun.Status.PipelineResults[0].Value.StringVal
				}
			}
		}
	}
	return nil
}
