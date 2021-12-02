package meteor

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/types"
	pointer "k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aicoe/meteor-operator/api/v1alpha1"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Submit a Tekton PipelineRun from a collection
func (r *MeteorReconciler) ReconcilePipelineRun(name string, ctx *context.Context, req ctrl.Request) error {
	res := &pipelinev1beta1.PipelineRun{}
	resourceName := fmt.Sprintf("%s-%s", r.Meteor.GetName(), name)
	namespacedName := types.NamespacedName{Name: resourceName, Namespace: req.NamespacedName.Namespace}

	logger := log.FromContext(*ctx).WithValues("pipelinerun", namespacedName)

	updateStatus := func(status metav1.ConditionStatus, reason, message string) {
		r.UpdateStatus(r.Meteor, "PipelineRun", name, status, reason, message)
	}

	statusIndex := func() int {
		for i, pr := range r.Meteor.Status.Pipelines {
			if pr.Name == name {
				return i
			}
		}
		result := v1alpha1.PipelineResult{
			Name:  name,
			Ready: "False",
		}
		r.Meteor.Status.Pipelines = append(r.Meteor.Status.Pipelines, result)
		return len(r.Meteor.Status.Pipelines) - 1
	}()

	if err := r.Get(*ctx, namespacedName, res); err != nil {
		if k8serrors.IsNotFound(err) {
			ownerReferences, err := r.ownerReferences()
			if err != nil || ownerReferences == "[]" {
				logger.Error(err, "Unable to serialize ownerReferences")
				return nil
			}
			logger.WithValues("ref", ownerReferences).Info("")

			logger.Info("Creating PipelineRun")

			res = &pipelinev1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: req.NamespacedName.Namespace,
				},
				Spec: pipelinev1beta1.PipelineRunSpec{
					PipelineRef: &pipelinev1beta1.PipelineRef{
						Name: name,
					},
					Params: []pipelinev1beta1.Param{
						{
							Name: "url",
							Value: pipelinev1beta1.ArrayOrString{
								Type:      pipelinev1beta1.ParamTypeString,
								StringVal: r.Meteor.Spec.Url,
							},
						},
						{
							Name: "ref",
							Value: pipelinev1beta1.ArrayOrString{
								Type:      pipelinev1beta1.ParamTypeString,
								StringVal: r.Meteor.Spec.Ref,
							},
						},
						{
							Name: "ownerReferences",
							Value: pipelinev1beta1.ArrayOrString{
								Type:      pipelinev1beta1.ParamTypeString,
								StringVal: string(ownerReferences),
							},
						},
					},
					Workspaces: []pipelinev1beta1.WorkspaceBinding{
						{
							Name: "data",
							VolumeClaimTemplate: &v1.PersistentVolumeClaim{
								Spec: v1.PersistentVolumeClaimSpec{
									AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
									Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											v1.ResourceStorage: resource.MustParse("500Mi"),
										},
									},
								},
							},
						},
						{
							Name: "sslcertdir",
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "openshift-service-ca.crt",
								},
								Items: []v1.KeyToPath{{
									Key:  "service-ca.crt",
									Path: "ca.crt",
								}},
								DefaultMode: pointer.Int32(420),
							},
						},
					},
				},
			}
			controllerutil.SetControllerReference(r.Meteor, res, r.Scheme)

			if len(r.Shower.Spec.Workspace.AccessModes) != 0 {
				res.Spec.Workspaces[0].VolumeClaimTemplate.Spec.AccessModes = r.Shower.Spec.Workspace.AccessModes
			}
			if !reflect.ValueOf(r.Shower.Spec.Workspace.Resources).IsZero() {
				res.Spec.Workspaces[0].VolumeClaimTemplate.Spec.Resources = r.Shower.Spec.Workspace.Resources
			}
			if r.Meteor.Spec.TTL == 0 && r.Shower.Spec.PersistentMeteorsHost != "" {
				res.Spec.Params = append(res.Spec.Params, pipelinev1beta1.Param{
					Name: "host",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      pipelinev1beta1.ParamTypeString,
						StringVal: r.Shower.Spec.PersistentMeteorsHost,
					},
				})
			}
			externalServices, err := r.externalServices()
			if err != nil {
				logger.Error(err, "Unable to serialize ownerReferences")
				return err
			} else {
				res.Spec.Params = append(res.Spec.Params, pipelinev1beta1.Param{
					Name: "externalServices",
					Value: pipelinev1beta1.ArrayOrString{
						Type:      pipelinev1beta1.ParamTypeString,
						StringVal: string(externalServices),
					},
				})
			}

			if err := r.Create(*ctx, res); err != nil {
				logger.Error(err, "Unable to create PipelineRun")
				updateStatus(metav1.ConditionTrue, "CreateError", fmt.Sprintf("Unable to create pipelinerun. %s", err))
				return err
			}
			updateStatus(metav1.ConditionTrue, "BuildStated", "Tekton pipeline was submitted.")
			return nil
		}
		logger.Error(err, "Error fetching PipelineRun")

		updateStatus(metav1.ConditionFalse, "Error", fmt.Sprintf("Reconcile resulted in error. %s", err))
		return err
	}

	if len(res.Status.Conditions) > 0 {
		if len(res.Status.Conditions) != 1 {
			logger.Error(nil, "Tekton reported multiple conditions")
		}
		condition := res.Status.Conditions[0]
		updateStatus(metav1.ConditionStatus(condition.Status), condition.Reason, condition.Message)
	}
	if res.Status.CompletionTime != nil && res.Status.Conditions[0].Reason == "Succeeded" {
		r.Meteor.Status.Pipelines[statusIndex].Ready = "True"
		if len(res.Status.PipelineResults) > 0 {
			r.Meteor.Status.Pipelines[statusIndex].Url = res.Status.PipelineResults[0].Value
		}
	}
	return nil
}

func (r *MeteorReconciler) ownerReferences() (string, error) {
	if len(r.Meteor.Status.Comas) == 0 {
		return "", errors.New("no Comas found")
	}
	allRefs := append(r.Meteor.Status.Comas, r.Meteor.GetReference(false))
	ownerReferences, err := json.CaseSensitiveJSONIterator().Marshal(allRefs)
	return string(ownerReferences), err
}

func (r *MeteorReconciler) externalServices() (string, error) {
	ownerReferences, err := json.CaseSensitiveJSONIterator().Marshal(r.Shower.Spec.ExternalServices)
	return string(ownerReferences), err
}
