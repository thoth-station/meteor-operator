package shower

import (
	"context"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ShowerReconciler) ReconcileRoute(ctx *context.Context, req ctrl.Request) error {
	res := &routev1.Route{}
	resourceName := fmt.Sprintf("meteor-shower-%s", r.Shower.GetName())
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.Namespace}

	desiredSpec := routev1.RouteSpec{
		Host: r.Shower.Spec.Ingress.Host,
		Path: r.Shower.Spec.Ingress.Path,
		To: routev1.RouteTargetReference{
			Kind: "Service",
			Name: resourceName,
		},
		Port: &routev1.RoutePort{
			TargetPort: intstr.FromString("http"),
		},
	}

	logger := log.FromContext(*ctx).WithValues("route", namespacedName)

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Creating")

			res = &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:        resourceName,
					Namespace:   req.NamespacedName.Namespace,
					Annotations: r.Shower.Spec.Ingress.Annotations,
					Labels:      r.Shower.Spec.Ingress.Labels,
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

	if !isMapASubset(res.ObjectMeta.Annotations, r.Shower.Spec.Ingress.Annotations) ||
		!isMapASubset(res.ObjectMeta.Labels, r.Shower.Spec.Ingress.Labels) ||
		res.Spec.Host != r.Shower.Spec.Ingress.Host ||
		res.Spec.Path != r.Shower.Spec.Ingress.Path {

		for k, v := range r.Shower.Spec.Ingress.Annotations {
			res.ObjectMeta.Annotations[k] = v
		}
		for k, v := range r.Shower.Spec.Ingress.Labels {
			res.ObjectMeta.Labels[k] = v
		}
		res.Spec.Host = r.Shower.Spec.Ingress.Host
		res.Spec.Path = r.Shower.Spec.Ingress.Path

		if err := r.Update(*ctx, res); err != nil {
			logger.Error(err, "Unable to update")
			return err
		}
	}

	logger.Info("Reconciled")
	return nil
}

func isMapASubset(superset, subset map[string]string) bool {
	for key, targetValue := range subset {
		if observedValue, ok := superset[key]; ok {
			if observedValue == targetValue {
				continue
			}
		}
		return false
	}
	return true
}
