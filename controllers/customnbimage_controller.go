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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
