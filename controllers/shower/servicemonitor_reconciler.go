package shower

import (
	"context"
	"fmt"
	"reflect"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ShowerReconciler) ReconcileServiceMonitor(ctx *context.Context, req ctrl.Request) error {
	res := &monitoringv1.ServiceMonitor{}
	resourceName := fmt.Sprintf("meteor-shower-%s", r.Shower.GetName())
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.Namespace}
	desiredSpec := monitoringv1.ServiceMonitorSpec{
		Selector: metav1.LabelSelector{
			MatchLabels: map[string]string{"shower.meteor.zone": resourceName},
		},
		Endpoints: []monitoringv1.Endpoint{
			{
				Port:   "http",
				Scheme: "http",
				Path:   "/metrics",
			},
		},
	}

	logger := log.FromContext(*ctx).WithValues("servicemonitor", namespacedName)

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Creating")

			res = &monitoringv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
				},
				Spec: desiredSpec,
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

	if !reflect.DeepEqual(res.Spec, desiredSpec) {
		res.Spec = desiredSpec
		if err := r.Update(*ctx, res); err != nil {
			logger.Error(err, "Unable to update")
			return err
		}
	}

	logger.Info("Reconciled")
	return nil
}
