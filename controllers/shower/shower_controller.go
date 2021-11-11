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

package shower

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aicoe/meteor-operator/api/v1alpha1"
	"github.com/aicoe/meteor-operator/version"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// ShowerReconciler reconciles a Shower object
type ShowerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Shower *v1alpha1.Shower
}

const (
	RequeueAfter     = 10 * time.Second
	DefaultImageBase = "quay.io/aicoe/meteor-shower"
)

//+kubebuilder:rbac:groups=meteor.zone,resources=showers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=meteor.zone,resources=showers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=meteor.zone,resources=showers/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Shower object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ShowerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	r.Shower = &v1alpha1.Shower{}
	if err := r.Get(ctx, req.NamespacedName, r.Shower); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Unable to fetch reconciled resource")
		return ctrl.Result{Requeue: true}, err
	}

	if !r.Shower.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Resource being delete, skipping further reconcile.")
		return ctrl.Result{}, nil
	}

	r.Shower.Status.ObservedGeneration = r.Shower.GetGeneration()

	if r.Shower.Spec.Image != "" {
		r.Shower.Status.Image = r.Shower.Spec.Image
	} else {
		r.Shower.Status.Image = DefaultImageBase + ":" + version.Version
	}

	actions := []func(*context.Context, reconcile.Request) error{
		r.ReconcileServiceAccount,
		r.ReconcileShowerRolebinding,
		r.ReconcileShowerRole,
		r.ReconcilePipelineRolebinding,
		r.ReconcilePipelineRole,
		r.ReconcileDeployment,
		r.ReconcileService,
		r.ReconcileServiceMonitor,
		r.ReconcileRoute,
	}
	for _, externalService := range r.Shower.Spec.ExternalServices {
		if externalService.Namespace != "" {
			actions = append(
				actions,
				r.ReconcileExternalRolebinding(externalService.Namespace),
				r.ReconcileExternalRole(externalService.Namespace),
			)
		}
	}

	for _, reconciler := range actions {
		if err := reconciler(&ctx, req); err != nil {
			return r.UpdateStatusNow(ctx, err)
		}
	}

	return ctrl.Result{}, nil
}

// Force object status update. Returns a reconcile result
func (r *ShowerReconciler) UpdateStatusNow(ctx context.Context, originalErr error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	if err := r.Status().Update(ctx, r.Shower); err != nil {
		logger.WithValues("reason", err.Error()).Info("Unable to update status, retrying")
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{RequeueAfter: RequeueAfter}, originalErr
}

// SetupWithManager sets up the controller with the Manager.
func (r *ShowerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Shower{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}
