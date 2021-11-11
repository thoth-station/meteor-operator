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

package meteor

import (
	"context"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aicoe/meteor-operator/api/v1alpha1"
	common "github.com/aicoe/meteor-operator/controllers/common"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const RequeueAfter = 10 * time.Second

// MeteorReconciler reconciles a Meteor object
type MeteorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Meteor *v1alpha1.Meteor
}

//+kubebuilder:rbac:groups=meteor.zone,resources=meteors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=meteor.zone,resources=meteors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=meteor.zone,resources=meteors/finalizers,verbs=update
//+kubebuilder:rbac:groups=meteor.zone,resources=meteorcomas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=meteor.zone,resources=meteorcomas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/finalizers,verbs=update

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
	r.Meteor = &v1alpha1.Meteor{}
	if err := r.Get(ctx, req.NamespacedName, r.Meteor); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Unable to fetch reconciled resource")
		return ctrl.Result{Requeue: true}, err
	}

	common.MetricsBeforeReconcile(r.Meteor)
	r.Meteor.Status.ObservedGeneration = r.Meteor.GetGeneration()
	r.Meteor.Status.Phase = r.Meteor.AggregatePhase()
	if r.Meteor.Spec.TTL != 0 {
		r.Meteor.Status.ExpirationTimestamp = metav1.NewTime(r.Meteor.GetExpirationTimestamp())
	}
	common.MetricsAfterReconcile(r.Meteor)

	if r.Meteor.IsTTLReached() && r.Meteor.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("TTL reached")
		if err := r.Delete(ctx, r.Meteor); err != nil {
			logger.Error(err, "Failed to delete")
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.EnsureFinalizers(ctx); err != nil {
		logger.Error(err, "Failed to ensure finalizers")
		return ctrl.Result{}, err
	}
	if !r.Meteor.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Resource being delete, skipping further reconcile.")
		return ctrl.Result{}, nil
	}

	if err := r.ReconcileComas(ctx); err != nil {
		return r.UpdateStatusNow(ctx, err)
	}

	for _, pipeline := range r.Meteor.Spec.Pipelines {
		if err := r.ReconcilePipelineRun(pipeline, &ctx, req); err != nil {
			return r.UpdateStatusNow(ctx, err)
		}
	}

	return r.UpdateStatusNow(ctx, nil)
}

// Force object status update. Returns a reconcile result
func (r *MeteorReconciler) UpdateStatusNow(ctx context.Context, originalErr error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	if err := r.Status().Update(ctx, r.Meteor); err != nil {
		logger.WithValues("reason", err.Error()).Info("Unable to update status, retrying")
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{RequeueAfter: RequeueAfter}, originalErr
}

// Set status condition helper
func (r *MeteorReconciler) UpdateStatus(meteor *v1alpha1.Meteor, kind, name string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&meteor.Status.Conditions, metav1.Condition{
		Type:               kind + strings.Title(name),
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: meteor.GetGeneration(),
	})
}

func (r *MeteorReconciler) EnsureFinalizers(ctx context.Context) error {
	logger := log.FromContext(ctx)

	finalizer := v1alpha1.GroupVersion.Group + "/finalizer"
	if r.Meteor.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, register our finalizer
		if !containsString(r.Meteor.GetFinalizers(), finalizer) {
			controllerutil.AddFinalizer(r.Meteor, finalizer)
			if err := r.Update(ctx, r.Meteor); err != nil {
				logger.Error(err, "Unable to add finalizer")
				return err
			}
		}
	} else {
		// The object is being deleted
		if containsString(r.Meteor.GetFinalizers(), finalizer) {
			if err := r.DeleteComas(ctx); err != nil {
				logger.Error(err, "Unable to delete Comas")
				return err
			}

			controllerutil.RemoveFinalizer(r.Meteor, finalizer)
			if err := r.Update(ctx, r.Meteor); err != nil {
				logger.Error(err, "Unable to remove finalizer")
				return err
			}
		}
	}
	return nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *MeteorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Meteor{}).
		Owns(&pipelinev1beta1.PipelineRun{}).
		Complete(r)
}
