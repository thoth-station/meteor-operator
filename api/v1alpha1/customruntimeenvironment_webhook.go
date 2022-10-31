/*
Copyright 2021, 2022 The Meteor Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var customruntimeenvironmentlog = logf.Log.WithName("customruntimeenvironment-resource")

func (r *CustomRuntimeEnvironment) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-meteor-zone-v1alpha1-customruntimeenvironment,mutating=true,failurePolicy=fail,sideEffects=None,groups=meteor.zone,resources=customruntimeenvironments,verbs=create;update,versions=v1alpha1,name=mcustomruntimeenvironment.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CustomRuntimeEnvironment{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CustomRuntimeEnvironment) Default() {
	customruntimeenvironmentlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

//+kubebuilder:webhook:path=/validate-meteor-zone-v1alpha1-customruntimeenvironment,mutating=false,failurePolicy=fail,sideEffects=None,groups=meteor.zone,resources=customruntimeenvironments,verbs=create;update,versions=v1alpha1,name=vcustomruntimeenvironment.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CustomRuntimeEnvironment{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CustomRuntimeEnvironment) ValidateCreate() error {
	customruntimeenvironmentlog.Info("validate create", "name", r.Name)

	return r.ValidateCRE()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CustomRuntimeEnvironment) ValidateUpdate(old runtime.Object) error {
	customruntimeenvironmentlog.Info("validate update", "name", r.Name)

	return r.ValidateCRE()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CustomRuntimeEnvironment) ValidateDelete() error {
	customruntimeenvironmentlog.Info("validate delete", "name", r.Name)

	// TODO change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
	// TODO fill in your validation logic upon object deletion.
	return nil
}

// ValidateCRE implements webhook.Validator for create/update
func (r *CustomRuntimeEnvironment) ValidateCRE() error {
	var allErrs field.ErrorList

	if err := r.validateCustomeRuntimeEnvironmentAnnotation(ODHNameAnnotationKey); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateCustomeRuntimeEnvironmentAnnotation(ODHDescriptionAnnotationKey); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateCustomeRuntimeEnvironmentAnnotation(ODHCreatorAnnotationKey); err != nil {
		allErrs = append(allErrs, err)
	}

	if r.Spec.BuildType == PackageList {
		if err := r.validateCustomeRuntimeEnvironmentPackageListBuildType(); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: Group, Kind: "CustomeRuntimeEnvironment"},
		r.Name, allErrs)
}

func (r *CustomRuntimeEnvironment) validateCustomeRuntimeEnvironmentAnnotation(annotation string) *field.Error {
	if r.Annotations == nil {
		return field.Required(field.NewPath("metadata.annotations"), "annotation is required")
	}

	if _, ok := r.Annotations[annotation]; !ok {
		return field.Required(field.NewPath("metadata.annotations").Key(annotation), "annotation is required")
	}

	return nil
}

func (r *CustomRuntimeEnvironment) validateCustomeRuntimeEnvironmentPackageListBuildType() *field.Error {
	logf.Log.Info("validateCustomeRuntimeEnvironmentPackageListBuildType", "r", r)

	if r.Spec.PackageVersions == nil {
		return field.Required(field.NewPath("spec.packageVersions"), "packageVersions is required")
	}

	if len(r.Spec.PackageVersions) == 0 {
		return field.Required(field.NewPath("spec.packageVersions"), "packageVersions is required")
	}

	if (r.Spec.BaseImage != "") && r.Spec.RuntimeEnvironment.isValid() {
		return field.Invalid(field.NewPath("spec.baseImage"), r.Spec.BaseImage, "baseImage and runtimeEnvironment are mutually exclusive")
	}

	if (r.Spec.BaseImage == "") && !r.Spec.RuntimeEnvironment.isValid() {
		return field.Required(field.NewPath("spec.baseImage"), "baseImage or runtimeEnvironment is required")
	}

	return nil
}
