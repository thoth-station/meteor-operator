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

package cre

import (
	"context"
	"fmt"
	"strconv"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
}

//+kubebuilder:rbac:groups=meteor.zone,resources=customruntimeenvironments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=meteor.zone,resources=customruntimeenvironments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=meteor.zone,resources=customruntimeenvironments/finalizers,verbs=update
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CustomRuntimeEnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)
	CRE := meteorv1alpha1.CustomRuntimeEnvironment{}

	if err := r.Get(ctx, req.NamespacedName, &CRE); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource was deleted.")
			err = nil
		}
		return ctrl.Result{}, err
	}

	if !CRE.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Resource being deleted, skipping further reconcile.")
		return ctrl.Result{}, nil
	}
	oldStatus := CRE.Status.DeepCopy()

	CRE.Status.Phase = CRE.AggregatePhase()

	// depending on the build type, we reconcile a pipelinerun
	r.reconcilePipelineRun(ctx, &CRE)

	// let's see if we can update the status
	CRE.Status.ObservedGeneration = CRE.Generation
	if !equality.Semantic.DeepEqual(CRE.Status, oldStatus) {
		logger.Info("Reconciled CustomNotebookImage", "spec", CRE.Spec, "status", CRE.Status)
		CRE.Status.Phase = CRE.AggregatePhase()
	}

	err := r.Status().Update(ctx, &CRE)
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

// reconcilePipelineRun will reconcile the pipeline run for the CustomRuntimeEnvironment.
func (r *CustomRuntimeEnvironmentReconciler) reconcilePipelineRun(ctx context.Context, cre *meteorv1alpha1.CustomRuntimeEnvironment) {

	build_types := map[meteorv1alpha1.BuildType]string{
		meteorv1alpha1.GitRepository: "gitrepo",
		meteorv1alpha1.PackageList:   "package-list",
		meteorv1alpha1.ImportImage:   "import",
	}

	pipeline := build_types[cre.Spec.BuildType]
	pipelineRun := &pipelinev1beta1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("cre-%s-%d-%s", cre.GetName(), cre.GetGeneration(), pipeline),
			Namespace: cre.Namespace,
			Labels: map[string]string{
				"cre.thoth-station.ninja/pipeline":         build_types[cre.Spec.BuildType],
				"cre.thoth-station.ninja/spouseGeneration": strconv.FormatInt(cre.GetGeneration(), 10),
			}}}
	namespacedName := types.NamespacedName{Name: pipelineRun.GetName(), Namespace: cre.Namespace}

	logger := log.FromContext(ctx).WithValues("pipelinerun", namespacedName)

	statusIndex := func() int {
		for i, pr := range cre.Status.Pipelines {
			if pr.Name == cre.Name {
				return i
			}
		}
		result := meteorv1alpha1.PipelineResult{
			Name:            cre.Name,
			Ready:           "False",
			PipelineRunName: pipelineRun.GetName(),
		}
		cre.Status.Pipelines = append(cre.Status.Pipelines, result)
		return len(cre.Status.Pipelines) - 1
	}()

	if err := r.Get(ctx, namespacedName, pipelineRun); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating PipelineRun")

			// let's put the mandatory annotations into the PipelineRun
			params := []pipelinev1beta1.Param{
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

			// Add the parameters specific to each build type
			switch buildType := cre.Spec.BuildTypeSpec.BuildType; buildType {
			case meteorv1alpha1.ImportImage:
				params = append(params, pipelinev1beta1.Param{
					Name: "baseImage",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      pipelinev1beta1.ParamTypeString,
						StringVal: cre.Spec.BuildTypeSpec.FromImage,
					}, // TODO we need a validator for this
				})
			case meteorv1alpha1.PackageList:
				// if we have a BaseImage supplied, use it
				if cre.Spec.BaseImage != "" {
					params = append(params, pipelinev1beta1.Param{
						Name: "baseImage",
						Value: pipelinev1beta1.ArrayOrString{
							Type:      pipelinev1beta1.ParamTypeString,
							StringVal: cre.Spec.BaseImage,
						},
					})
				} else {
					params = append(params, pipelinev1beta1.Param{
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
					params = append(params, pipelinev1beta1.Param{
						Name: "packages",
						Value: pipelinev1beta1.ArrayOrString{
							Type:     pipelinev1beta1.ParamTypeArray,
							ArrayVal: cre.Spec.PackageVersions,
						},
					})
				}
			case meteorv1alpha1.GitRepository:
				params = append(params, pipelinev1beta1.Param{
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

			pipelineRun = &pipelinev1beta1.PipelineRun{
				ObjectMeta: pipelineRun.ObjectMeta,
				Spec: pipelinev1beta1.PipelineRunSpec{
					PipelineRef: &pipelinev1beta1.PipelineRef{
						Name: fmt.Sprintf("cre-%s", pipeline),
					},
					Params:     params,
					Workspaces: workspaces_const,
				},
			}
			controllerutil.SetControllerReference(cre, pipelineRun, r.Scheme)

			if err := r.Create(ctx, pipelineRun); err != nil {
				logger.Error(err, "Unable to create PipelineRun")

				meta.SetStatusCondition(&cre.Status.Conditions,
					metav1.Condition{
						ObservedGeneration: cre.Generation,
						Type:               meteorv1alpha1.ErrorPipelineRunCreate,
						Status:             metav1.ConditionTrue,
						Reason:             "PipelineRunCreateFailed",
						Message:            err.Error(),
					})
			}
			logger.Info("Created PipelineRun for CNBI", "PipelineRun", pipelineRun.GetNamespacedName(), "CRE", cre)
			meta.SetStatusCondition(&cre.Status.Conditions, metav1.Condition{
				ObservedGeneration: cre.Generation,
				Type:               meteorv1alpha1.PipelineRunCreated,
				Status:             metav1.ConditionTrue,
				Reason:             "PipelineRunCreated",
				Message:            fmt.Sprintf("%s PipelineRun created successfully", pipelineRun.Name),
			})
			return
		}

		logger.Error(err, "Error fetching PipelineRun")
		meta.SetStatusCondition(&cre.Status.Conditions, metav1.Condition{
			ObservedGeneration: cre.Generation,
			Type:               meteorv1alpha1.GenericPipelineError,
			Status:             metav1.ConditionTrue,
			Reason:             "PipelineRunGenericError",
			Message:            err.Error()},
		)
		return
	}

	if len(pipelineRun.Status.Conditions) > 0 {
		if len(pipelineRun.Status.Conditions) != 1 { // TODO observe tekton project if they stay with just one condition all the time
			logger.Error(nil, "Tekton reported multiple conditions")
		}

		// Let's check if the PipelineRun is completed successfully or not, and conclude our new conditions
		if pipelineRun.Status.Conditions[0].Status == v1.ConditionTrue && pipelineRun.Status.Conditions[0].Type == "Succeeded" {
			meta.SetStatusCondition(&cre.Status.Conditions, metav1.Condition{
				ObservedGeneration: cre.Generation,
				Type:               meteorv1alpha1.PipelineRunCompleted,
				Status:             metav1.ConditionTrue,
				Reason:             "PipelineRunCompleted",
				Message:            "The PipelineRun has been completed successfully.",
			})
			meta.RemoveStatusCondition(&cre.Status.Conditions, meteorv1alpha1.PipelineRunCreated)

			if pipelineRun.Labels["cre.thoth-station.ninja/pipeline"] == "import" {
				meta.SetStatusCondition(&cre.Status.Conditions, metav1.Condition{
					ObservedGeneration: cre.Generation,
					Type:               meteorv1alpha1.ImageImportReady,
					Status:             metav1.ConditionTrue,
					Reason:             "ImageImportReady",
					Message:            "Import succeeded, the image is ready to be used.",
				})
			}
			if pipelineRun.Labels["cre.thoth-station.ninja/pipeline"] == "package-list" {
				meta.SetStatusCondition(&cre.Status.Conditions, metav1.Condition{
					ObservedGeneration: cre.Generation,
					Type:               meteorv1alpha1.PackageListBuildCompleted,
					Status:             metav1.ConditionTrue,
					Reason:             "PackageListBuildCompleted",
					Message:            "Build from Package List succeeded, the image is ready to be used.",
				})
			}

			// TODO add other pipeline-specific success conditions
		} else if pipelineRun.Status.Conditions[0].Status == v1.ConditionFalse && pipelineRun.Status.Conditions[0].Type == "Succeeded" {
			meta.SetStatusCondition(&cre.Status.Conditions, metav1.Condition{
				ObservedGeneration: cre.Generation,
				Type:               meteorv1alpha1.PipelineRunCompleted,
				Status:             metav1.ConditionTrue,
				Reason:             "PipelineRunCompleted",
				Message:            "The PipelineRun has been completed with a failure!",
			})
			meta.RemoveStatusCondition(&cre.Status.Conditions, meteorv1alpha1.PipelineRunCreated)

			if pipelineRun.Labels["cre.thoth-station.ninja/pipeline"] == "import" {
				meta.SetStatusCondition(&cre.Status.Conditions, metav1.Condition{
					ObservedGeneration: cre.Generation,
					Type:               meteorv1alpha1.ImageImportReady,
					Status:             metav1.ConditionFalse,
					Reason:             "ImageImportNotReady",
					Message:            "Import failed, this could be due to the repository to import from does not exist or is not accessible",
				})
			} else if pipelineRun.Labels["cre.thoth-station.ninja/pipeline"] == "gitrepo" {
				meta.SetStatusCondition(&cre.Status.Conditions, metav1.Condition{
					ObservedGeneration: cre.Generation,
					Type:               meteorv1alpha1.ErrorBuildingImage,
					Status:             metav1.ConditionTrue,
					Reason:             "ErrorBuildingImage",
					Message:            "Build failed!",
				})
			} else if pipelineRun.Labels["cre.thoth-station.ninja/pipeline"] == "package-list" {
				meta.SetStatusCondition(&cre.Status.Conditions, metav1.Condition{
					ObservedGeneration: cre.Generation,
					Type:               meteorv1alpha1.PackageListBuildCompleted,
					Status:             metav1.ConditionFalse,
					Reason:             "PackageListBuildCompleted",
					Message:            "Build from Package List failed!",
				})
			}

		}

		return

	}

	if pipelineRun.Status.CompletionTime != nil {
		if pipelineRun.Status.Conditions[0].Reason == "Succeeded" { // FIXME this feels dangerous, is it really the 1st one? all the time?
			cre.Status.Pipelines[statusIndex].Ready = "True"
			if len(pipelineRun.Status.PipelineResults) > 0 {
				if pipelineRun.Status.PipelineResults[0].Value.Type == pipelinev1beta1.ParamTypeString {
					cre.Status.Pipelines[statusIndex].Url = pipelineRun.Status.PipelineResults[0].Value.StringVal
				}
			}
		} else if pipelineRun.Status.Conditions[0].Reason == "Failed" {
			cre.Status.Pipelines[statusIndex].Ready = "False"
		}
	}

	return
}
