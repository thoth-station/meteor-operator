package shower

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ShowerReconciler) ReconcileDeployment(ctx *context.Context, req ctrl.Request) error {
	res := &appsv1.Deployment{}
	resourceName := fmt.Sprintf("meteor-shower-%s", r.Shower.GetName())
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("deployment", namespacedName)

	desiredSpec := appsv1.DeploymentSpec{
		Replicas: &r.Shower.Spec.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"shower.meteor.zone": resourceName},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"shower.meteor.zone": resourceName},
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: resourceName,
				Containers: []corev1.Container{
					{
						Name:            "shower",
						Image:           r.Shower.Status.Image,
						ImagePullPolicy: corev1.PullAlways,
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("300Mi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("300Mi"),
							},
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 3000,
							},
						},
						Env: r.Shower.Spec.Env,
					},
				},
			},
		},
	}

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Creating ServiceAccount")

			res = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
				},
				Spec: desiredSpec,
			}
			controllerutil.SetControllerReference(r.Shower, res, r.Scheme)

			if err := r.Create(*ctx, res); err != nil {
				logger.Error(err, "Unable to create ServiceAccount")
				return err
			}
			return nil
		}
		logger.Error(err, "Error fetching Deployment")
		return err
	}

	if res.Spec.Replicas != desiredSpec.Replicas || !reflect.DeepEqual(res.Spec.Selector, desiredSpec.Selector) || !reflect.DeepEqual(res.Spec.Template, desiredSpec.Template) {
		res.Spec.Replicas = desiredSpec.Replicas
		res.Spec.Selector = desiredSpec.Selector
		res.Spec.Template = desiredSpec.Template
		if err := r.Update(*ctx, res); err != nil {
			logger.Error(err, "Error reconciling")
			return err
		}
	}
	return nil
}
