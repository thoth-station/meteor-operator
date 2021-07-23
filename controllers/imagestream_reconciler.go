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

func updateImageStreamStatus(meteor *meteorv1alpha1.Meteor, name string, status metav1.ConditionStatus, reason, message string) {
	updateStatus(meteor, "ImageStream", name, status, reason, message)
}
func (r *MeteorReconciler) ReconcileImageStream(name string, ctx *context.Context, req ctrl.Request, meteor *meteorv1alpha1.Meteor, status *meteorv1alpha1.MeteorImage) error {
	logger := log.FromContext(*ctx)

	imagestream := &imagev1.ImageStream{}
	imagestreamName := "meteor-" + string(meteor.GetUID())
	imageName := GetImageName(req.Namespace, name, meteor.GetUID())
	imagestreamLabels := MeteorLabels(meteor)
	imagestreamLabels["opendatahub.io/notebook-image"] = "true"

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

	if err := r.Get(*ctx, types.NamespacedName{Name: imagestreamName, Namespace: req.NamespacedName.Namespace}, imagestream); err != nil {
		if errors.IsNotFound(err) {
			imagestream = &imagev1.ImageStream{
				ObjectMeta: metav1.ObjectMeta{
					Name:      imagestreamName,
					Namespace: req.NamespacedName.Namespace,
					Labels:    imagestreamLabels,
					Annotations: map[string]string{
						"opendatahub.io/notebook-image-url":  meteor.Spec.Url,
						"opendatahub.io/notebook-image-name": meteor.GetName(),
						"opendatahub.io/notebook-image-desc": fmt.Sprintf("JupyterHub image for Meteor %s", imagestreamName),
					},
				},
				Spec: *newSpec,
			}
			controllerutil.SetControllerReference(meteor, imagestream, r.Scheme)

			if err := r.Create(*ctx, imagestream); err != nil {
				logger.Error(err, fmt.Sprintf("Unable to create imagestream %s", imagestreamName))
				updateImageStreamStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to create imagestream. %s", err))
				return err
			}

			logger.Info(fmt.Sprintf("Imagestream '%s' created.", imagestreamName))
			updateImageStreamStatus(meteor, name, metav1.ConditionTrue, "Created", "Imagestream was created.")
			return nil
		}
		logger.Error(err, fmt.Sprintf("Error fetching '%s' imagestream.", imagestreamName))
		updateImageStreamStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	// FIXME More carefull compare is required
	// if !reflect.DeepEqual(imagestream.Spec, *newSpec) {
	// 	imagestream.Spec = *newSpec
	// 	if err := r.Update(*ctx, imagestream); err != nil {
	// 		logger.Error(err, fmt.Sprintf("Unable to update imagestream %s", imagestreamName))
	// 		updateImageStreamStatus(meteor, name, metav1.ConditionFalse, "Error", fmt.Sprintf("Unable to update imagestream. %s", err))
	// 		return err
	// 	}
	// 	logger.Info(fmt.Sprintf("Imagestream '%s' updated.", imagestreamName))
	// }
	status.ImageStreamName = imagestreamName
	updateImageStreamStatus(meteor, name, metav1.ConditionTrue, "Ready", "Imagestream was reconciled successfully.")
	return nil
}
