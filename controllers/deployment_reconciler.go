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
	logger := log.FromContext(*ctx)
	deployment := &appsv1.Deployment{}
	deploymentName := meteor.GetName()
	deploymentLabels := MeteorLabels(meteor)
	newSpec := &appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: deploymentLabels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: deploymentLabels,
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

	if err := r.Get(*ctx, types.NamespacedName{Name: deploymentName, Namespace: req.NamespacedName.Namespace}, deployment); err != nil {
		if errors.IsNotFound(err) {
			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: req.NamespacedName.Namespace,
					Labels:    deploymentLabels,
				},
				Spec: *newSpec,
			}
			controllerutil.SetControllerReference(meteor, deployment, r.Scheme)

			if err := r.Create(*ctx, deployment); err != nil {
				logger.Error(err, fmt.Sprintf("Unable to create deployment %s", deploymentName))
				updateDeploymentStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to create deployment. %s", err))
				return err
			}

			logger.Info(fmt.Sprintf("Deployment '%s' created.", deploymentName))
			updateDeploymentStatus(meteor, name, metav1.ConditionTrue, "Created", "Deployment was created.")
			return nil
		}
		logger.Error(err, fmt.Sprintf("Error fetching '%s' deployment.", deploymentName))

		updateDeploymentStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if !reflect.DeepEqual(deployment.Spec, *newSpec) {
		deployment.Spec = *newSpec
		if err := r.Update(*ctx, deployment); err != nil {
			logger.Error(err, fmt.Sprintf("Unable to update deployment %s", deploymentName))
			updateDeploymentStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to update deployment. %s", err))
			return err
		}
		logger.Info(fmt.Sprintf("Deployment '%s' updated.", deploymentName))
	}
	updateDeploymentStatus(meteor, name, metav1.ConditionTrue, "Ready", "Deployment was reconciled successfully.")
	return nil
}
