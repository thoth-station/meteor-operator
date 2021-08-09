package controllers

import (
	"context"
	"fmt"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *MeteorReconciler) ReconcileImageStream(name string, ctx *context.Context, req ctrl.Request, status *meteorv1alpha1.MeteorImage) error {
	res := &imagev1.ImageStream{}
	resourceName := fmt.Sprintf("%s-%s", r.Meteor.GetName(), name)
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("imagestream", namespacedName)

	imageName := fmt.Sprintf(ImageFormatter, req.Namespace, r.Meteor.GetName(), name)
	labels := r.Meteor.SeedLabels()
	labels[meteorv1alpha1.ODHJupyterHubLabel] = "true"

	newSpec := &imagev1.ImageStreamSpec{
		LookupPolicy: imagev1.ImageLookupPolicy{Local: true},
		Tags: []imagev1.TagReference{
			{
				Name: "latest",
				Annotations: map[string]string{
					"openshift.io/imported-from": imageName,
				},
				From: &v1.ObjectReference{
					Kind: "DockerImage",
					Name: imageName,
				},
				ImportPolicy: imagev1.TagImportPolicy{
					Scheduled: true,
				},
			},
		},
	}
	updateStatus := func(status metav1.ConditionStatus, reason, message string) {
		r.UpdateStatus(r.Meteor, "ImageStream", name, status, reason, message)
	}

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating ImageStream")

			res = &imagev1.ImageStream{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
					Labels:    labels,
					Annotations: map[string]string{
						"opendatahub.io/notebook-image-url":  r.Meteor.Spec.Url,
						"opendatahub.io/notebook-image-name": r.Meteor.GetName(),
						"opendatahub.io/notebook-image-desc": fmt.Sprintf("JupyterHub image for Meteor %s", resourceName),
					},
				},
				Spec: *newSpec,
			}
			controllerutil.SetControllerReference(r.Meteor, res, r.Scheme)

			if err := r.Create(*ctx, res); err != nil {
				logger.Error(err, "Unable to create ImageStream")
				updateStatus(metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to create ImageStream. %s", err))
				return err
			}
			updateStatus(metav1.ConditionTrue, "Created", "ImageStream was created.")
			return nil
		}
		logger.Error(err, "Error fetching ImageStream")
		updateStatus(metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	// FIXME More carefull compare is required
	// if !reflect.DeepEqual(imagestream.Spec, *newSpec) {
	// 	imagestream.Spec = *newSpec
	// 	if err := r.Update(*ctx, imagestream); err != nil {
	// 		logger.Error(err, "Unable to update ImageStream %s")
	// 		updateStatus(metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to update imagestream. %s", err))
	// 		return err
	// 	}
	// }
	status.ImageStreamName = resourceName
	updateStatus(metav1.ConditionTrue, "Ready", "Imagestream was reconciled successfully.")
	return nil
}
