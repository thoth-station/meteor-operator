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

	// depending on the build type, we reconcile a pipelinerun
	newStatus := &meteorv1alpha1.CustomNotebookImageStatus{}

	switch buildType := r.CNBi.Spec.BuildTypeSpec.BuildType; buildType {
	case meteorv1alpha1.ImportImage:
		newStatus = r.reconcilePipelineRun("import", ctx, req)
	case meteorv1alpha1.GitRepository:
		newStatus = r.reconcilePipelineRun("gitrepo", ctx, req)
	case meteorv1alpha1.PackageList:
		newStatus = r.reconcilePipelineRun("package-list", ctx, req)
	}

	// let's see if we can update the status
	newStatus.ObservedGeneration = r.CNBi.Generation
	if equality.Semantic.DeepEqual(newStatus, &r.CNBi.Status) {
		return r.UpdateStatusNow(ctx, newStatus, nil)
	}

	/* TODO check if the PipelineRun ran for the current runtime environment
	 * if not, delete the PipelineRun and reconcile?
	 */

	logger.Info("Reconciled CustomNotebookImage", "spec", r.CNBi.Spec, "status", r.CNBi.Status)
	r.CNBi.Status.Phase = r.CNBi.AggregatePhase()
	err := r.updateStatus(ctx, req.NamespacedName, newStatus)
	return ctrl.Result{}, err
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
func (r *CustomNBImageReconciler) UpdateStatusNow(ctx context.Context, status *meteorv1alpha1.CustomNotebookImageStatus, originalErr error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	status.DeepCopyInto(&r.CNBi.Status)
	logger.Info(("Updating status of CustomNotebookImage"), "status", r.CNBi.Status, "newStatus", status)

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

func (r *CustomNBImageReconciler) updateStatus(ctx context.Context, nn types.NamespacedName, status *meteorv1alpha1.CustomNotebookImageStatus) error {
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		original := &meteorv1alpha1.CustomNBImage{}
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

// reconcilePipelineRun will reconcile the pipeline run for the CustomNBImage.
func (r *CustomNBImageReconciler) reconcilePipelineRun(name string, ctx context.Context, req ctrl.Request) *meteorv1alpha1.CustomNotebookImageStatus {
	pipelineRun := &pipelinev1beta1.PipelineRun{}
	resourceName := fmt.Sprintf("cnbi-%s-%s", r.CNBi.GetName(), name)
	pipelineReferenceName := fmt.Sprintf("cnbi-%s", name)
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	newStatus := r.CNBi.Status.DeepCopy()

	logger := log.FromContext(ctx).WithValues("pipelinerun", namespacedName)

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

	if err := r.Get(ctx, namespacedName, pipelineRun); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating PipelineRun")

			// let's put the mandatory annotations into the PipelineRun
			params := []pipelinev1beta1.Param{
				{
					Name: "name",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      pipelinev1beta1.ParamTypeString,
						StringVal: r.CNBi.ObjectMeta.Annotations["opendatahub.io/notebook-image-name"],
					},
				},
				{
					Name: "creator",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      pipelinev1beta1.ParamTypeString,
						StringVal: r.CNBi.ObjectMeta.Annotations["opendatahub.io/notebook-image-creator"],
					},
				},
				{
					Name: "description",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      pipelinev1beta1.ParamTypeString,
						StringVal: r.CNBi.ObjectMeta.Annotations["opendatahub.io/notebook-image-desc"],
					},
				},
			}

			// Add the parameters specific to each build type
			switch buildType := r.CNBi.Spec.BuildTypeSpec.BuildType; buildType {
			case meteorv1alpha1.ImportImage:
				params = append(params, pipelinev1beta1.Param{
					Name: "baseImage",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      pipelinev1beta1.ParamTypeString,
						StringVal: r.CNBi.Spec.BuildTypeSpec.FromImage,
					}, // TODO we need a validator for this
				})
			case meteorv1alpha1.PackageList:
				// if we have a BaseImage supplied, use it
				if r.CNBi.Spec.BaseImage != "" {
					params = append(params, pipelinev1beta1.Param{
						Name: "baseImage",
						Value: pipelinev1beta1.ArrayOrString{
							Type:      pipelinev1beta1.ParamTypeString,
							StringVal: r.CNBi.Spec.BaseImage,
						},
					})
				} else {
					params = append(params, pipelinev1beta1.Param{
						Name: "osVersion",
						Value: pipelinev1beta1.ArrayOrString{
							Type:      pipelinev1beta1.ParamTypeString,
							StringVal: r.CNBi.Spec.RuntimeEnvironment.OSVersion,
						},
					}, pipelinev1beta1.Param{
						Name: "osName",
						Value: pipelinev1beta1.ArrayOrString{
							Type:      pipelinev1beta1.ParamTypeString,
							StringVal: r.CNBi.Spec.RuntimeEnvironment.OSName,
						},
					}, pipelinev1beta1.Param{
						Name: "pythonVersion",
						Value: pipelinev1beta1.ArrayOrString{
							Type:      pipelinev1beta1.ParamTypeString,
							StringVal: r.CNBi.Spec.RuntimeEnvironment.PythonVersion,
						},
					})
				}

				// if we have no PackageVersion specified, we are done...
				if len(r.CNBi.Spec.PackageVersions) > 0 {
					params = append(params, pipelinev1beta1.Param{
						Name: "packages",
						Value: pipelinev1beta1.ArrayOrString{
							Type:     pipelinev1beta1.ParamTypeArray,
							ArrayVal: r.CNBi.Spec.PackageVersions,
						},
					})
				}
			case meteorv1alpha1.GitRepository:
				params = append(params, pipelinev1beta1.Param{
					Name: "url",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      pipelinev1beta1.ParamTypeString,
						StringVal: r.CNBi.Spec.BuildTypeSpec.Repository,
					},
				}, pipelinev1beta1.Param{
					Name: "ref",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      pipelinev1beta1.ParamTypeString,
						StringVal: r.CNBi.Spec.BuildTypeSpec.GitRef,
					},
				})
			}

			pipelineRun = &pipelinev1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
					Labels: map[string]string{
						"cnbi.thoth-station.ninja/pipeline": name,
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
			controllerutil.SetControllerReference(r.CNBi, pipelineRun, r.Scheme)

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
		if len(pipelineRun.Status.Conditions) != 1 {
			logger.Error(nil, "Tekton reported multiple conditions")
		}

		if pipelineRun.Labels["cnbi.thoth-station.ninja/pipeline"] == "import" {
			if pipelineRun.Status.Conditions[0].Status == v1.ConditionFalse && pipelineRun.Status.Conditions[0].Type == "Succeeded" {
				logger.Info("Import pipeline failed")
				setCondition(newStatus, meteorv1alpha1.ImageImportReady, metav1.ConditionFalse, "ImageImportNotReady", "Import failed, this could be due to the repository to import from does not exist or is not accessible")
				setCondition(newStatus, meteorv1alpha1.PipelineRunCompleted, metav1.ConditionTrue, "PipelineRunCompleted", "The PipelineRun has been completed, Image is importend")
				removeCondition(newStatus, meteorv1alpha1.PipelineRunCreated)
			}
		}

		return newStatus

	}

	if pipelineRun.Status.CompletionTime != nil {
		if pipelineRun.Status.Conditions[0].Reason == "Succeeded" { // FIXME this feels dangerous, is it really the 1st one? all the time?
			r.CNBi.Status.Pipelines[statusIndex].Ready = "True"
			if len(pipelineRun.Status.PipelineResults) > 0 {
				if pipelineRun.Status.PipelineResults[0].Value.Type == pipelinev1beta1.ParamTypeString {
					r.CNBi.Status.Pipelines[statusIndex].Url = pipelineRun.Status.PipelineResults[0].Value.StringVal
				}
			}
		} else if pipelineRun.Status.Conditions[0].Reason == "Failed" {
			r.CNBi.Status.Pipelines[statusIndex].Ready = "False"
		}
	}

	return newStatus
}
