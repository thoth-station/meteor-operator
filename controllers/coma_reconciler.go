package controllers

import (
	"context"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *MeteorReconciler) ReconcileComas(ctx context.Context) error {
	logger := log.FromContext(ctx)
	namespaces := []string{"opf-jupyterhub"}
	name := r.Meteor.GetName()

	for _, namespace := range namespaces {
		coma := &meteorv1alpha1.MeteorComa{}
		namespacedName := types.NamespacedName{Name: r.Meteor.GetName(), Namespace: namespace}

		if err := r.Get(ctx, namespacedName, coma); err != nil {
			if errors.IsNotFound(err) {
				coma = &meteorv1alpha1.MeteorComa{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
				}
				if err := r.Create(ctx, coma); err != nil {
					logger.Error(err, "Unable to create Coma")
					return err
				}
			}
		}

		if coma.APIVersion == "" || coma.UID == "" {
			// Coma was not processed by Kube api yet, wait for next event
			return nil
		}

		ref := meteorv1alpha1.NamespacedOwnerReference{
			OwnerReference: *metav1.NewControllerRef(coma, coma.GroupVersionKind()),
			Namespace:      namespace,
		}
		isControlled := false
		ref.Controller = &isControlled
		if !containsComa(r.Meteor.Status.Comas, ref) {
			r.Meteor.Status.Comas = append(r.Meteor.Status.Comas, ref)
		}

		coma.Status.Owner = r.Meteor.GetReference(true)

		if err := r.Status().Update(ctx, coma); err != nil {
			logger.Error(err, "Unable to update Coma status")
		}
	}
	return nil
}

func (r *MeteorReconciler) DeleteComas(ctx context.Context) error {
	logger := log.FromContext(ctx)
	for _, coma := range r.Meteor.Status.Comas {
		comaMeta := &meteorv1alpha1.MeteorComa{
			ObjectMeta: metav1.ObjectMeta{Name: coma.Name, Namespace: coma.Namespace},
		}
		logger.WithValues("coma", comaMeta).Info("Deleting coma")
		if err := r.Delete(ctx, comaMeta); err != nil {
			logger.WithValues("coma", comaMeta).Error(err, "Failed to delete coma")
			return err
		}
	}
	return nil
}

func containsComa(slice []meteorv1alpha1.NamespacedOwnerReference, ref meteorv1alpha1.NamespacedOwnerReference) bool {
	for _, item := range slice {
		if item.Namespace == ref.Namespace && item.Name == ref.Name && item.UID == ref.UID {
			return true
		}
	}
	return false
}
