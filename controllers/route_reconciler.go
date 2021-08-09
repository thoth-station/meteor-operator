package controllers

import (
	"context"
	"fmt"
	"reflect"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func updateRouteStatus(meteor *meteorv1alpha1.Meteor, name string, status metav1.ConditionStatus, reason, message string) {
	updateStatus(meteor, "Route", name, status, reason, message)
}
func (r *MeteorReconciler) ReconcileRoute(name string, ctx *context.Context, req ctrl.Request, meteor *meteorv1alpha1.Meteor, status *meteorv1alpha1.MeteorImage) error {
	res := &routev1.Route{}
	resourceName := meteor.GetName()
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("route", namespacedName)

	labels := MeteorLabels(meteor)
	newSpec := &routev1.RouteSpec{
		To: routev1.RouteTargetReference{Kind: "Service", Name: meteor.GetName()},
	}

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating Route")
			res = &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
					Labels:    labels,
				},
				Spec: *newSpec,
			}
			controllerutil.SetControllerReference(meteor, res, r.Scheme)

			if err := r.Create(*ctx, res); err != nil {
				logger.Error(err, "Unable to create Route")
				updateRouteStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to create route. %s", err))
				return err
			}

			updateRouteStatus(meteor, name, metav1.ConditionTrue, "Created", "Route was created.")
			return nil
		}
		logger.Error(err, "Error fetching Route.")
		updateRouteStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if !reflect.DeepEqual(res.Spec, *newSpec) {
		res.Spec = *newSpec
		if err := r.Update(*ctx, res); err != nil {
			logger.Error(err, "Unable to update Route")
			updateRouteStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to update route. %s", err))
			return err
		}
	}
	status.Url = res.Spec.Host
	updateRouteStatus(meteor, name, metav1.ConditionTrue, "Ready", "Route was reconciled successfully.")
	return nil
}
