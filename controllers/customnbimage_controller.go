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
	"time"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	// depending on the import Strategy, we reconcile a pipelinerun
	// Check for the PipelineRun reconcilation, and update the status of the CustomNBImage resource
	if r.CNBi.Spec.BuildTypeSpec.BuildType == meteorv1alpha1.ImportImage {
		if err := r.ReconcilePipelineRun("import", &ctx, req); err != nil {
			return r.UpdateStatusNow(ctx, err)
		}
	} else {
		// we assume its a build from UI parameters...

		if err := r.ReconcilePipelineRun("prepare", &ctx, req); err != nil {
			return r.UpdateStatusNow(ctx, err)
		}
	}

	/* TODO check if the PipelineRun ran for the current runtime environment
	 * if not, delete the PipelineRun and reconcile?
	 */

	logger.Info("Reconciled CustomNBImage", "spec", r.CNBi.Spec)
	return r.UpdateStatusNow(ctx, nil)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomNBImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO setup index for PipelineRuns

	return ctrl.NewControllerManagedBy(mgr).
		For(&meteorv1alpha1.CustomNBImage{}).
		Owns(&pipelinev1beta1.PipelineRun{}).
		Owns(&meteorv1alpha1.Meteor{}).
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
		Type:    kind + cases.Title(language.Und).String(name),
		Status:  status,
		Reason:  reason,
		Message: message,
	})
}

// ReconcilePipelineRun will reconcile the pipeline run for the CustomNBImage.
func (r *CustomNBImageReconciler) ReconcilePipelineRun(name string, ctx *context.Context, req ctrl.Request) error {
	pipelineRun := &pipelinev1beta1.PipelineRun{}
	resourceName := fmt.Sprintf("cnbi-%s-%s", r.CNBi.GetName(), name)
	pipelineReferenceName := fmt.Sprintf("cnbi-%s", name)
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
		result := meteorv1alpha1.PipelineResult{
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

			// let's put the mandatory annotations into the PipelineRun
			params := []pipelinev1beta1.Param{
				{
					Name: "name",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      "string",
						StringVal: r.CNBi.ObjectMeta.Annotations["opendatahub.io/notebook-image-name"],
					},
				},
				{
					Name: "creator",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      "string",
						StringVal: r.CNBi.ObjectMeta.Annotations["opendatahub.io/notebook-image-creator"],
					},
				},
				{
					Name: "description",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      "string",
						StringVal: r.CNBi.ObjectMeta.Annotations["opendatahub.io/notebook-image-desc"],
					},
				},
			}

			if r.CNBi.Spec.BuildTypeSpec.BuildType == meteorv1alpha1.ImportImage {
				params = append(params, pipelinev1beta1.Param{
					Name: "baseImage",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      "string",
						StringVal: r.CNBi.Spec.BuildTypeSpec.FromImage,
					}, // TODO we need a validator for this
				})
			}

			if r.CNBi.Spec.BuildTypeSpec.BuildType == meteorv1alpha1.PackageList {

				// if we have a BaseImage supplied, use it
				if r.CNBi.Spec.RuntimeEnvironment.BuilderImage != "" {
					params = append(params, pipelinev1beta1.Param{
						Name: "baseImage",
						Value: pipelinev1beta1.ArrayOrString{
							Type:      pipelinev1beta1.ParamTypeString,
							StringVal: r.CNBi.Spec.RuntimeEnvironment.BuilderImage,
						},
					})
				} else {
					params = append(params, pipelinev1beta1.Param{
						Name: "osVersion",
						Value: pipelinev1beta1.ArrayOrString{
							Type:      pipelinev1beta1.ParamTypeString,
							StringVal: r.CNBi.Spec.RuntimeEnvironment.OSVersion,
						},
					})
					params = append(params, pipelinev1beta1.Param{
						Name: "osName",
						Value: pipelinev1beta1.ArrayOrString{
							Type:      pipelinev1beta1.ParamTypeString,
							StringVal: r.CNBi.Spec.RuntimeEnvironment.OSName,
						},
					})
					params = append(params, pipelinev1beta1.Param{
						Name: "pythonVersion",
						Value: pipelinev1beta1.ArrayOrString{
							Type:      pipelinev1beta1.ParamTypeString,
							StringVal: r.CNBi.Spec.RuntimeEnvironment.PythonVersion,
						},
					})
				}

				// if we have no PackageVersion specified, we are done...
				if len(r.CNBi.Spec.PackageVersion) > 0 {
					params = append(params, pipelinev1beta1.Param{
						Name: "packages",
						Value: pipelinev1beta1.ArrayOrString{
							Type:     pipelinev1beta1.ParamTypeArray,
							ArrayVal: r.CNBi.Spec.PackageVersion,
						},
					})
				}
			}

			pipelineRun = &pipelinev1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
				},
				Spec: pipelinev1beta1.PipelineRunSpec{
					PipelineRef: &pipelinev1beta1.PipelineRef{
						Name: pipelineReferenceName,
					},
					Params: params,
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
				updateStatus(metav1.ConditionTrue, "ErrorPipelineRunCreate", fmt.Sprintf("Unable to create PipelineRun. %s", err))
				return err
			}
			updateStatus(metav1.ConditionTrue, "SuccessPipelineRunCreate", "Tekton PipelineRun was created.")
			return nil
		}
		logger.Error(err, "Error fetching PipelineRun")

		updateStatus(metav1.ConditionFalse, "ErrorPipelineRun", fmt.Sprintf("Reconcile resulted in error. %s", err))
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
