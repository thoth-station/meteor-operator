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

func (r *MeteorReconciler) ReconcileRoute(name string, ctx *context.Context, req ctrl.Request, status *meteorv1alpha1.MeteorImage) error {
	res := &routev1.Route{}
	resourceName := r.Meteor.GetName()
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("route", namespacedName)

	labels := r.Meteor.SeedLabels()
	newSpec := &routev1.RouteSpec{
		To: routev1.RouteTargetReference{Kind: "Service", Name: r.Meteor.GetName()},
	}

	updateStatus := func(status metav1.ConditionStatus, reason, message string) {
		r.UpdateStatus(r.Meteor, "Route", name, status, reason, message)
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
			controllerutil.SetControllerReference(r.Meteor, res, r.Scheme)

			if err := r.Create(*ctx, res); err != nil {
				logger.Error(err, "Unable to create Route")
				updateStatus(metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to create route. %s", err))
				return err
			}

			updateStatus(metav1.ConditionTrue, "Created", "Route was created.")
			return nil
		}
		logger.Error(err, "Error fetching Route.")
		updateStatus(metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if !reflect.DeepEqual(res.Spec, *newSpec) {
		res.Spec = *newSpec
		if err := r.Update(*ctx, res); err != nil {
			logger.Error(err, "Unable to update Route")
			updateStatus(metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to update route. %s", err))
			return err
		}
	}
	status.Url = res.Spec.Host
	updateStatus(metav1.ConditionTrue, "Ready", "Route was reconciled successfully.")
	return nil
}
