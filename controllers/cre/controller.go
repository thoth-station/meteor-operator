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
	"context"
	"fmt"
	"time"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	meteorv1alpha1 "github.com/thoth-station/meteor-operator/api/v1alpha1"
)

// CustomRuntimeEnvironmentReconciler reconciles a CustomRuntimeEnvironment object
type CustomRuntimeEnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	CRE    *meteorv1alpha1.CustomRuntimeEnvironment
}

//+kubebuilder:rbac:groups=meteor.zone,resources=customtruntimeenvironment,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=meteor.zone,resources=customtruntimeenvironment/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=meteor.zone,resources=customtruntimeenvironment/finalizers,verbs=update
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CustomRuntimeEnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	r.CRE = &meteorv1alpha1.CustomRuntimeEnvironment{}

	if err := r.Get(ctx, req.NamespacedName, r.CRE); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Unable to fetch reconciled resource")
		return ctrl.Result{Requeue: true}, err
	}

	if !r.CRE.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Resource being delete, skipping further reconcile.")
		return ctrl.Result{}, nil
	}

	r.CRE.Status.Phase = r.CRE.AggregatePhase()

	// depending on the build type, we reconcile a pipelinerun
	newStatus := &meteorv1alpha1.CustomRuntimeEnvironmentStatus{}

	switch buildType := r.CRE.Spec.BuildTypeSpec.BuildType; buildType {
	case meteorv1alpha1.ImportImage:
		newStatus = r.reconcilePipelineRun("import", ctx, req)
	case meteorv1alpha1.GitRepository:
		newStatus = r.reconcilePipelineRun("gitrepo", ctx, req)
	case meteorv1alpha1.PackageList:
		newStatus = r.reconcilePipelineRun("package-list", ctx, req)
	}

	// let's see if we can update the status
	newStatus.ObservedGeneration = r.CRE.Generation
	if equality.Semantic.DeepEqual(newStatus, &r.CRE.Status) {
		return r.UpdateStatusNow(ctx, newStatus, nil)
	}

	/* TODO check if the PipelineRun ran for the current runtime environment
	 * if not, delete the PipelineRun and reconcile?
	 */

	logger.Info("Reconciled CustomNotebookImage", "spec", r.CRE.Spec, "status", r.CRE.Status)
	r.CRE.Status.Phase = r.CRE.AggregatePhase()
	err := r.updateStatus(ctx, req.NamespacedName, newStatus)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomRuntimeEnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO setup index for PipelineRuns

	return ctrl.NewControllerManagedBy(mgr).
		For(&meteorv1alpha1.CustomRuntimeEnvironment{}).
		Owns(&pipelinev1beta1.PipelineRun{}).
		Owns(&meteorv1alpha1.Meteor{}).
		Complete(r)
}

// Force object status update. Returns a reconcile result
func (r *CustomRuntimeEnvironmentReconciler) UpdateStatusNow(ctx context.Context, status *meteorv1alpha1.CustomRuntimeEnvironmentStatus, originalErr error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	status.DeepCopyInto(&r.CRE.Status)
	logger.Info(("Updating status of CustomNotebookImage"), "status", r.CRE.Status, "newStatus", status)

	if err := r.Status().Update(ctx, r.CRE); err != nil {
		logger.WithValues("reason", err.Error()).Info("Unable to update status, retrying")
		return ctrl.Result{Requeue: true}, nil
	}
	if originalErr != nil {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, originalErr
	} else {
		return ctrl.Result{}, nil
	}
}

func (r *CustomRuntimeEnvironmentReconciler) updateStatus(ctx context.Context, nn types.NamespacedName, status *meteorv1alpha1.CustomRuntimeEnvironmentStatus) error {
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		original := &meteorv1alpha1.CustomRuntimeEnvironment{}
		if err := r.Get(ctx, nn, original); err != nil {
			return err
		}
		original.Status = *status
		if err := r.Client.Status().Update(ctx, original); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to update status of Application %s/%s: %v", nn.Namespace, nn.Name, err)
	}
	return nil
}

// reconcilePipelineRun will reconcile the pipeline run for the CRE
func (r *CustomRuntimeEnvironmentReconciler) reconcilePipelineRun(name string, ctx context.Context, req ctrl.Request) *meteorv1alpha1.CustomRuntimeEnvironmentStatus {
	pipelineRun := &pipelinev1beta1.PipelineRun{}
	resourceName := fmt.Sprintf("cre-%s-%s", r.CRE.GetName(), name)
	pipelineReferenceName := fmt.Sprintf("cre-%s", name)
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	newStatus := r.CRE.Status.DeepCopy()

	logger := log.FromContext(ctx).WithValues("pipelinerun", namespacedName)

	statusIndex := func() int {
		for i, pr := range r.CRE.Status.Pipelines {
			if pr.Name == name {
				return i
			}
		}
		result := meteorv1alpha1.PipelineResult{
			Name:            name,
			Ready:           "False",
			PipelineRunName: resourceName,
		}
		r.CRE.Status.Pipelines = append(r.CRE.Status.Pipelines, result)
		return len(r.CRE.Status.Pipelines) - 1
	}()

	if err := r.Get(ctx, namespacedName, pipelineRun); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating PipelineRun")

			params := []pipelinev1beta1.Param{}

			// let's put the mandatory annotations into the PipelineRun
			params = appendODHAnnotations(r.CRE, params)

			// append the parameters specific to each build type
			params = appendBuildTypeParameters(r.CRE, params)

			generatePipelineRun(r.CRE, pipelineRun, pipelineReferenceName, name, namespacedName, params)
			_ = controllerutil.SetControllerReference(r.CRE, pipelineRun, r.Scheme) // TODO: check error

			if err := r.Create(ctx, pipelineRun); err != nil {
				logger.Error(err, "Unable to create PipelineRun")

				setCondition(newStatus, meteorv1alpha1.ErrorPipelineRunCreate, metav1.ConditionTrue, "PipelineRunCreateFailed", err.Error())
				return newStatus
			}
			setCondition(newStatus, meteorv1alpha1.PipelineRunCreated, metav1.ConditionTrue, "PipelineRunCreated", fmt.Sprintf("%s PipelineRun created successfully", name))
			return newStatus
		}

		logger.Error(err, "Error fetching PipelineRun")
		setCondition(newStatus, meteorv1alpha1.GenericPipelineError, metav1.ConditionTrue, "PipelineRunGenericError", err.Error())
		return newStatus
	}

	if len(pipelineRun.Status.Conditions) > 0 {
		if len(pipelineRun.Status.Conditions) != 1 { // TODO observe tekton project if they stay with just one condition all the time
			logger.Error(nil, "Tekton reported multiple conditions")
		}

		if pipelineRun.Labels["cre.thoth-station.ninja/pipeline"] == "import" {
			if pipelineRun.Status.Conditions[0].Status == v1.ConditionFalse && pipelineRun.Status.Conditions[0].Type == "Succeeded" {
				setCondition(newStatus, meteorv1alpha1.ImageImportReady, metav1.ConditionFalse, "ImageImportNotReady", "Import failed, this could be due to the repository to import from does not exist or is not accessible")
				setCondition(newStatus, meteorv1alpha1.PipelineRunCompleted, metav1.ConditionTrue, "PipelineRunCompleted", "The PipelineRun has been completed, but the Image could not be imported!")
				removeCondition(newStatus, meteorv1alpha1.PipelineRunCreated)
			} else if pipelineRun.Status.Conditions[0].Status == v1.ConditionTrue && pipelineRun.Status.Conditions[0].Type == "Succeeded" {
				setCondition(newStatus, meteorv1alpha1.ImageImportReady, metav1.ConditionTrue, "ImageImportReady", "Import succeeded, the image is ready to be used")
				setCondition(newStatus, meteorv1alpha1.PipelineRunCompleted, metav1.ConditionTrue, "PipelineRunCompleted", "The PipelineRun has been completed, Image is importend")
				removeCondition(newStatus, meteorv1alpha1.PipelineRunCreated)
			}
		}

		return newStatus

	}

	if pipelineRun.Status.CompletionTime != nil {
		if pipelineRun.Status.Conditions[0].Reason == "Succeeded" { // FIXME this feels dangerous, is it really the 1st one? all the time?
			r.CRE.Status.Pipelines[statusIndex].Ready = "True"
			if len(pipelineRun.Status.PipelineResults) > 0 {
				if pipelineRun.Status.PipelineResults[0].Value.Type == pipelinev1beta1.ParamTypeString {
					r.CRE.Status.Pipelines[statusIndex].Url = pipelineRun.Status.PipelineResults[0].Value.StringVal
				}
			}
		} else if pipelineRun.Status.Conditions[0].Reason == "Failed" {
			r.CRE.Status.Pipelines[statusIndex].Ready = "False"
		}
	}

	return newStatus
}

func generatePipelineRun(cre *meteorv1alpha1.CustomRuntimeEnvironment, pipelineRun *pipelinev1beta1.PipelineRun, pipelineReferenceName, name string, namespacedName types.NamespacedName, params []pipelinev1beta1.Param) {
	_pipelineRun := &pipelinev1beta1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
			Labels: map[string]string{
				"cre.thoth-station.ninja/pipeline": name,
			},
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

	_pipelineRun.DeepCopyInto(pipelineRun)
}

func appendODHAnnotations(cre *meteorv1alpha1.CustomRuntimeEnvironment, params []pipelinev1beta1.Param) []pipelinev1beta1.Param {
	_params := []pipelinev1beta1.Param{
		{
			Name: "name",
			Value: pipelinev1beta1.ArrayOrString{
				Type:      pipelinev1beta1.ParamTypeString,
				StringVal: cre.ObjectMeta.Annotations["opendatahub.io/notebook-image-name"],
			},
		},
		{
			Name: "creator",
			Value: pipelinev1beta1.ArrayOrString{
				Type:      pipelinev1beta1.ParamTypeString,
				StringVal: cre.ObjectMeta.Annotations["opendatahub.io/notebook-image-creator"],
			},
		},
		{
			Name: "description",
			Value: pipelinev1beta1.ArrayOrString{
				Type:      pipelinev1beta1.ParamTypeString,
				StringVal: cre.ObjectMeta.Annotations["opendatahub.io/notebook-image-desc"],
			},
		},
	}

	return append(params, _params...)
}

func appendBuildTypeParameters(cre *meteorv1alpha1.CustomRuntimeEnvironment, params []pipelinev1beta1.Param) []pipelinev1beta1.Param {
	_params := []pipelinev1beta1.Param{}

	switch buildType := cre.Spec.BuildTypeSpec.BuildType; buildType {
	case meteorv1alpha1.ImportImage:
		_params = append(_params, pipelinev1beta1.Param{
			Name: "baseImage",
			Value: pipelinev1beta1.ArrayOrString{
				Type:      pipelinev1beta1.ParamTypeString,
				StringVal: cre.Spec.BuildTypeSpec.FromImage,
			}, // TODO we need a validator for this
		})
	case meteorv1alpha1.PackageList:
		// if we have a BaseImage supplied, use it
		if cre.Spec.BaseImage != "" {
			_params = append(_params, pipelinev1beta1.Param{
				Name: "baseImage",
				Value: pipelinev1beta1.ArrayOrString{
					Type:      pipelinev1beta1.ParamTypeString,
					StringVal: cre.Spec.BaseImage,
				},
			})
		} else {
			_params = append(_params, pipelinev1beta1.Param{
				Name: "osVersion",
				Value: pipelinev1beta1.ArrayOrString{
					Type:      pipelinev1beta1.ParamTypeString,
					StringVal: cre.Spec.RuntimeEnvironment.OSVersion,
				},
			}, pipelinev1beta1.Param{
				Name: "osName",
				Value: pipelinev1beta1.ArrayOrString{
					Type:      pipelinev1beta1.ParamTypeString,
					StringVal: cre.Spec.RuntimeEnvironment.OSName,
				},
			}, pipelinev1beta1.Param{
				Name: "pythonVersion",
				Value: pipelinev1beta1.ArrayOrString{
					Type:      pipelinev1beta1.ParamTypeString,
					StringVal: cre.Spec.RuntimeEnvironment.PythonVersion,
				},
			})
		}

		// if we have no PackageVersion specified, we are done...
		if len(cre.Spec.PackageVersions) > 0 {
			_params = append(_params, pipelinev1beta1.Param{
				Name: "packages",
				Value: pipelinev1beta1.ArrayOrString{
					Type:     pipelinev1beta1.ParamTypeArray,
					ArrayVal: cre.Spec.PackageVersions,
				},
			})
		}
	case meteorv1alpha1.GitRepository:
		_params = append(_params, pipelinev1beta1.Param{
			Name: "url",
			Value: pipelinev1beta1.ArrayOrString{
				Type:      pipelinev1beta1.ParamTypeString,
				StringVal: cre.Spec.BuildTypeSpec.Repository,
			},
		}, pipelinev1beta1.Param{
			Name: "ref",
			Value: pipelinev1beta1.ArrayOrString{
				Type:      pipelinev1beta1.ParamTypeString,
				StringVal: cre.Spec.BuildTypeSpec.GitRef,
			},
		})
	}

	return append(params, _params...)
}
