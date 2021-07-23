package controllers

import (
	"context"
	"fmt"
	"reflect"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func updateServiceStatus(meteor *meteorv1alpha1.Meteor, name string, status metav1.ConditionStatus, reason, message string) {
	updateStatus(meteor, "Service", name, status, reason, message)
}

func (r *MeteorReconciler) ReconcileService(name string, ctx *context.Context, req ctrl.Request, meteor *meteorv1alpha1.Meteor) error {
	logger := log.FromContext(*ctx)
	service := &v1.Service{}
	serviceName := "meteor-" + string(meteor.GetUID())
	newSpec := &v1.ServiceSpec{
		Ports:    []v1.ServicePort{{Name: "http", Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080), Port: 8080}},
		Selector: MeteorLabels(meteor),
	}

	if err := r.Get(*ctx, types.NamespacedName{Name: serviceName, Namespace: req.NamespacedName.Namespace}, service); err != nil {
		if errors.IsNotFound(err) {
			service = &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceName,
					Namespace: req.NamespacedName.Namespace,
				},
				Spec: *newSpec,
			}

			controllerutil.SetControllerReference(meteor, service, r.Scheme)
			if err := r.Create(*ctx, service); err != nil {
				logger.Error(err, fmt.Sprintf("Unable to create service %s", serviceName))
				updateServiceStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to create service. %s", err))
				return err
			}

			logger.Info(fmt.Sprintf("Service '%s' created.", serviceName))
			updateServiceStatus(meteor, name, metav1.ConditionTrue, "Created", "Service was created.")
			return nil
		}
		logger.Error(err, fmt.Sprintf("Error fetching '%s' service.", serviceName))
		updateServiceStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if !reflect.DeepEqual(service.Spec.Selector, newSpec.Selector) || !reflect.DeepEqual(service.Spec.Ports, newSpec.Ports) {
		service.Spec = *newSpec
		if err := r.Update(*ctx, service); err != nil {
			logger.Error(err, fmt.Sprintf("Unable to update service %s", serviceName))
			updateServiceStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to update service. %s", err))
			return err
		}
		logger.Info(fmt.Sprintf("Service '%s' updated.", serviceName))

	}
	updateServiceStatus(meteor, name, metav1.ConditionTrue, "Ready", "Service was reconciled successfully.")
	return nil
}
