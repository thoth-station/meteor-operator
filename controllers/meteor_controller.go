/*
Copyright 2021.

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

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MeteorReconciler reconciles a Meteor object
type MeteorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const MeteorPipelineAnnotation = "meteor.operate-first.cloud/pipeline"

// Submit a Tekton PipelineRun from a collection
func createPipelineRun(ctx *context.Context, meteor *meteorv1alpha1.Meteor, r *MeteorReconciler, pipelineName string) {
	logger := log.FromContext(*ctx)
	pipelineRun := &pipelinev1beta1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("meteor-%s-%s", pipelineName, meteor.UID),
			Annotations: map[string]string{
				MeteorPipelineAnnotation: pipelineName,
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

	if err := r.Client.Create(*ctx, pipelineRun); err != nil {
		logger.Error(err, fmt.Sprintf("Unable to instantiate %s pipeline", pipelineName))
	}

	meta.SetStatusCondition(&meteor.Status.Conditions, metav1.Condition{
		Type:               pipelineName,
		Status:             "True",
		Reason:             "BuildStated",
		Message:            "Tekton pipeline was submitted.",
		ObservedGeneration: meteor.GetGeneration(),
	})
}

//+kubebuilder:rbac:groups=meteor.operate-first.cloud,resources=meteors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=meteor.operate-first.cloud,resources=meteors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=meteor.operate-first.cloud,resources=meteors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Meteor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *MeteorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	meteor := &meteorv1alpha1.Meteor{}
	if err := r.Get(ctx, req.NamespacedName, meteor); err != nil {
		logger.Error(err, "Unable to fetch reconciled resource.")
	}

	if meteor.ReadyToStartPipeline("jupyterBook") {
		createPipelineRun(&ctx, meteor, r, "jupyterBook")
	}
	if meteor.ReadyToStartPipeline("jupyterHub") {
		createPipelineRun(&ctx, meteor, r, "jupyterHub")
	}

	meteor.Status.ObservedGeneration = meteor.GetGeneration()
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MeteorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&meteorv1alpha1.Meteor{}).
		Owns(&pipelinev1beta1.PipelineRun{}).
		Complete(r)
}
