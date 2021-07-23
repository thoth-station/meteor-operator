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
	logger := log.FromContext(*ctx)
	route := &routev1.Route{}
	routeName := meteor.InterpolateResourceName(meteorv1alpha1.Route)
	routeLabels := MeteorLabels(meteor)
	newSpec := &routev1.RouteSpec{
		To: routev1.RouteTargetReference{Kind: "Service", Name: meteor.InterpolateResourceName(meteorv1alpha1.Service)},
	}

	if err := r.Get(*ctx, types.NamespacedName{Name: routeName, Namespace: req.NamespacedName.Namespace}, route); err != nil {
		if errors.IsNotFound(err) {
			route = &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routeName,
					Namespace: req.NamespacedName.Namespace,
					Labels:    routeLabels,
				},
				Spec: *newSpec,
			}
			controllerutil.SetControllerReference(meteor, route, r.Scheme)

			if err := r.Create(*ctx, route); err != nil {
				logger.Error(err, fmt.Sprintf("Unable to create route %s", routeName))
				updateRouteStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to create route. %s", err))
				return err
			}

			logger.Info(fmt.Sprintf("Route '%s' created.", routeName))
			updateRouteStatus(meteor, name, metav1.ConditionTrue, "Created", "Route was created.")
			return nil
		}
		logger.Error(err, fmt.Sprintf("Error fetching '%s' route.", routeName))
		updateRouteStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if !reflect.DeepEqual(route.Spec, *newSpec) {
		route.Spec = *newSpec
		if err := r.Update(*ctx, route); err != nil {
			logger.Error(err, fmt.Sprintf("Unable to update route %s", routeName))
			updateRouteStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to update route. %s", err))
			return err
		}
		logger.Info(fmt.Sprintf("Route '%s' updated.", routeName))
	}
	status.Url = route.Spec.Host
	updateRouteStatus(meteor, name, metav1.ConditionTrue, "Ready", "Route was reconciled successfully.")
	return nil
}
