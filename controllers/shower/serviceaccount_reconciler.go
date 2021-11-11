package shower

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ShowerReconciler) ReconcileServiceAccount(ctx *context.Context, req ctrl.Request) error {
	res := &corev1.ServiceAccount{}
	resourceName := fmt.Sprintf("meteor-shower-%s", r.Shower.GetName())
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("serviceaccount", namespacedName)

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Creating")

			res = &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
				},
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
	logger.Info("Reconciled")
	return nil
}
