package shower

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aicoe/meteor-operator/api/v1alpha1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ShowerReconciler) reconcileRole(resourceName, namespace string, desiredRules []rbacv1.PolicyRule, ctx *context.Context, req ctrl.Request) error {
	res := &rbacv1.Role{}
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: namespace}

	logger := log.FromContext(*ctx).WithValues("role", namespacedName)

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Creating")

			res = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
				},
				Rules: desiredRules,
			}
			controllerutil.SetControllerReference(r.Shower, res, r.Scheme)

			if err := r.Create(*ctx, res); err != nil {
				logger.Error(err, "Unable to create")
				return err
			}
			logger.Info("Created")
			return nil
		}
		logger.Error(err, "Error fetching resource")
		return err
	}

	if !reflect.DeepEqual(res.Rules, desiredRules) {
		res.Rules = desiredRules
		if err := r.Update(*ctx, res); err != nil {
			logger.Error(err, "Unable to update")
			return err
		}
	}

	logger.Info("Reconciled")
	return nil
}

func (r *ShowerReconciler) ReconcileShowerRole(ctx *context.Context, req ctrl.Request) error {
	resourceName := fmt.Sprintf("meteor-shower-%s", r.Shower.GetName())
	desiredRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{v1alpha1.GroupVersion.Group},
			Resources: []string{"meteors"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{v1alpha1.GroupVersion.Group},
			Resources: []string{"meteors/status"},
			Verbs:     []string{"get"},
		},
	}
	return r.reconcileRole(resourceName, req.Namespace, desiredRules, ctx, req)
}

func (r *ShowerReconciler) ReconcilePipelineRole(ctx *context.Context, req ctrl.Request) error {
	resourceName := fmt.Sprintf("meteor-pipeline-%s", r.Shower.GetName())
	desiredRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{pipelinev1beta1.SchemeGroupVersion.Group},
			Resources: []string{"pipelineruns/finalizers"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{v1alpha1.GroupVersion.Group},
			Resources: []string{"meteors/finalizers"},
			Verbs:     []string{"update"},
		},
	}
	return r.reconcileRole(resourceName, req.Namespace, desiredRules, ctx, req)
}

func (r *ShowerReconciler) ReconcileExternalRole(namespace string) func(*context.Context, ctrl.Request) error {
	return func(ctx *context.Context, req ctrl.Request) error {
		resourceName := fmt.Sprintf("meteor-external-%s-%s", req.Namespace, r.Shower.GetName())

		desiredRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{imagev1.SchemeGroupVersion.Group},
				Resources: []string{"imagestreams", "imagestreams/layers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{routev1.SchemeGroupVersion.Group},
				Resources: []string{"routes", "routes/status"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{v1alpha1.GroupVersion.Group},
				Resources: []string{"meteorcomas/finalizers"},
				Verbs:     []string{"update"},
			},
		}
		return r.reconcileRole(resourceName, namespace, desiredRules, ctx, req)
	}
}
