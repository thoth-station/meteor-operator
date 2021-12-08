package shower

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ShowerReconciler) ReconcileService(ctx *context.Context, req ctrl.Request) error {
	res := &corev1.Service{}
	resourceName := fmt.Sprintf("meteor-shower-%s", r.Shower.GetName())
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.Namespace}
	desiredSpec := corev1.ServiceSpec{
		Selector: getSelector(resourceName),
		Type:     corev1.ServiceTypeClusterIP,
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       3000,
				TargetPort: intstr.FromInt(3000),
			},
		},
	}

	logger := log.FromContext(*ctx).WithValues("service", namespacedName)

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Creating")

			res = &corev1.Service{
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

	if !reflect.DeepEqual(res.Spec.Selector, desiredSpec.Selector) || res.Spec.Type != desiredSpec.Type || !reflect.DeepEqual(res.Spec.Ports, desiredSpec.Ports) {
		res.Spec.Selector = desiredSpec.Selector
		res.Spec.Type = desiredSpec.Type
		res.Spec.Ports = desiredSpec.Ports
		if err := r.Update(*ctx, res); err != nil {
			logger.Error(err, "Unable to update")
			return err
		}
	}
	return nil
}
