package shower

import (
	"context"
	"fmt"
	"reflect"

	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ShowerReconciler) reconcileRolebinding(resourceName, namespace string, desiredSubjects []rbacv1.Subject, desiredRoleRef rbacv1.RoleRef, ctx *context.Context, req ctrl.Request) error {
	res := &rbacv1.RoleBinding{}
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: namespace}

	logger := log.FromContext(*ctx).WithValues("rolebinding", namespacedName)

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Creating")

			res = &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Subjects: desiredSubjects,
				RoleRef:  desiredRoleRef,
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

	if !reflect.DeepEqual(res.Subjects, desiredSubjects) || !reflect.DeepEqual(res.RoleRef, desiredRoleRef) {
		res.Subjects = desiredSubjects
		res.RoleRef = desiredRoleRef

		if err := r.Update(*ctx, res); err != nil {
			logger.Error(err, "Unable to update")
			return err
		}
	}
	return nil
}

func (r *ShowerReconciler) ReconcileShowerRolebinding(ctx *context.Context, req ctrl.Request) error {
	resourceName := fmt.Sprintf("meteor-shower-%s", r.Shower.GetName())
	desiredSubjects := []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      resourceName,
			Namespace: req.Namespace,
		},
	}
	desiredRoleRef := rbacv1.RoleRef{
		APIGroup: rbacv1.SchemeGroupVersion.Group,
		Kind:     "Role",
		Name:     resourceName,
	}
	return r.reconcileRolebinding(resourceName, req.Namespace, desiredSubjects, desiredRoleRef, ctx, req)
}

func (r *ShowerReconciler) ReconcilePipelineRolebinding(ctx *context.Context, req ctrl.Request) error {
	resourceName := fmt.Sprintf("meteor-pipeline-%s", r.Shower.GetName())
	desiredSubjects := []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "pipeline",
			Namespace: req.Namespace,
		},
	}
	desiredRoleRef := rbacv1.RoleRef{
		APIGroup: rbacv1.SchemeGroupVersion.Group,
		Kind:     "Role",
		Name:     resourceName,
	}
	return r.reconcileRolebinding(resourceName, req.Namespace, desiredSubjects, desiredRoleRef, ctx, req)
}

func (r *ShowerReconciler) ReconcileExternalRolebinding(namespace string) func(*context.Context, ctrl.Request) error {
	return func(ctx *context.Context, req ctrl.Request) error {
		resourceName := fmt.Sprintf("meteor-external-%s-%s", req.Namespace, r.Shower.GetName())
		desiredSubjects := []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "pipeline",
				Namespace: req.Namespace,
			},
		}
		desiredRoleRef := rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     "Role",
			Name:     resourceName,
		}
		return r.reconcileRolebinding(resourceName, namespace, desiredSubjects, desiredRoleRef, ctx, req)
	}
}
