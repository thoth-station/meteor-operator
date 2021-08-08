package controllers

import (
	"strings"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(meteor *meteorv1alpha1.Meteor, kind, name string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&meteor.Status.Conditions, metav1.Condition{
		Type:               kind + strings.Title(name),
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: meteor.GetGeneration(),
	})
}

func MeteorLabels(meteor *meteorv1alpha1.Meteor) map[string]string {
	return map[string]string{MeteorLabel: string(meteor.GetUID())}
}
