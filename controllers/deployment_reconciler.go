package controllers

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"
)

func updateDeploymentStatus(meteor *meteorv1alpha1.Meteor, name string, status metav1.ConditionStatus, reason, message string) {
	updateStatus(meteor, "Deployment", name, status, reason, message)
}

// Reconciler for a deployment owned from Metor
func (r *MeteorReconciler) ReconcileDeployment(name string, ctx *context.Context, req ctrl.Request, meteor *meteorv1alpha1.Meteor) error {
	res := &appsv1.Deployment{}
	resourceName := meteor.GetName()
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("deployment", namespacedName)

	labels := MeteorLabels(meteor)
	newSpec := &appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  name,
						Image: fmt.Sprintf("%s-%s", meteor.GetName(), name),
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("100m"),
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
		},
	}

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating Deployment")
			res = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
					Labels:    labels,
				},
				Spec: *newSpec,
			}
			controllerutil.SetControllerReference(meteor, res, r.Scheme)

			if err := r.Create(*ctx, res); err != nil {
				logger.Error(err, "Unable to create Deployment")
				updateDeploymentStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to create deployment. %s", err))
				return err
			}

			updateDeploymentStatus(meteor, name, metav1.ConditionTrue, "Created", "Deployment was created.")
			return nil
		}
		logger.Error(err, "Error fetching Deployment")

		updateDeploymentStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if !reflect.DeepEqual(res.Spec, *newSpec) {
		res.Spec = *newSpec
		if err := r.Update(*ctx, res); err != nil {
			logger.Error(err, "Unable to update Deployment")
			updateDeploymentStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to update deployment. %s", err))
			return err
		}
	}
	updateDeploymentStatus(meteor, name, metav1.ConditionTrue, "Ready", "Deployment was reconciled successfully.")
	return nil
}
