//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BuildTypeSpec) DeepCopyInto(out *BuildTypeSpec) {
	*out = *in
	out.ImagePullSecret = in.ImagePullSecret
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BuildTypeSpec.
func (in *BuildTypeSpec) DeepCopy() *BuildTypeSpec {
	if in == nil {
		return nil
	}
	out := new(BuildTypeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Coma) DeepCopyInto(out *Coma) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Coma.
func (in *Coma) DeepCopy() *Coma {
	if in == nil {
		return nil
	}
	out := new(Coma)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Coma) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComaList) DeepCopyInto(out *ComaList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Coma, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComaList.
func (in *ComaList) DeepCopy() *ComaList {
	if in == nil {
		return nil
	}
	out := new(ComaList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ComaList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComaSpec) DeepCopyInto(out *ComaSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComaSpec.
func (in *ComaSpec) DeepCopy() *ComaSpec {
	if in == nil {
		return nil
	}
	out := new(ComaSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComaStatus) DeepCopyInto(out *ComaStatus) {
	*out = *in
	in.Owner.DeepCopyInto(&out.Owner)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComaStatus.
func (in *ComaStatus) DeepCopy() *ComaStatus {
	if in == nil {
		return nil
	}
	out := new(ComaStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentStatus) DeepCopyInto(out *ComponentStatus) {
	*out = *in
	if in.Running != nil {
		in, out := &in.Running, &out.Running
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Succeeded != nil {
		in, out := &in.Succeeded, &out.Succeeded
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Failed != nil {
		in, out := &in.Failed, &out.Failed
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComponentStatus.
func (in *ComponentStatus) DeepCopy() *ComponentStatus {
	if in == nil {
		return nil
	}
	out := new(ComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Conditions) DeepCopyInto(out *Conditions) {
	{
		in := &in
		*out = make(Conditions, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Conditions.
func (in Conditions) DeepCopy() Conditions {
	if in == nil {
		return nil
	}
	out := new(Conditions)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomRuntimeEnvironment) DeepCopyInto(out *CustomRuntimeEnvironment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomRuntimeEnvironment.
func (in *CustomRuntimeEnvironment) DeepCopy() *CustomRuntimeEnvironment {
	if in == nil {
		return nil
	}
	out := new(CustomRuntimeEnvironment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomRuntimeEnvironment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomRuntimeEnvironmentList) DeepCopyInto(out *CustomRuntimeEnvironmentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomRuntimeEnvironment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomRuntimeEnvironmentList.
func (in *CustomRuntimeEnvironmentList) DeepCopy() *CustomRuntimeEnvironmentList {
	if in == nil {
		return nil
	}
	out := new(CustomRuntimeEnvironmentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomRuntimeEnvironmentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomRuntimeEnvironmentRuntimeSpec) DeepCopyInto(out *CustomRuntimeEnvironmentRuntimeSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomRuntimeEnvironmentRuntimeSpec.
func (in *CustomRuntimeEnvironmentRuntimeSpec) DeepCopy() *CustomRuntimeEnvironmentRuntimeSpec {
	if in == nil {
		return nil
	}
	out := new(CustomRuntimeEnvironmentRuntimeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomRuntimeEnvironmentSpec) DeepCopyInto(out *CustomRuntimeEnvironmentSpec) {
	*out = *in
	out.RuntimeEnvironment = in.RuntimeEnvironment
	if in.PackageVersions != nil {
		in, out := &in.PackageVersions, &out.PackageVersions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.BuildTypeSpec = in.BuildTypeSpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomRuntimeEnvironmentSpec.
func (in *CustomRuntimeEnvironmentSpec) DeepCopy() *CustomRuntimeEnvironmentSpec {
	if in == nil {
		return nil
	}
	out := new(CustomRuntimeEnvironmentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomRuntimeEnvironmentStatus) DeepCopyInto(out *CustomRuntimeEnvironmentStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Pipelines != nil {
		in, out := &in.Pipelines, &out.Pipelines
		*out = make([]PipelineResult, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomRuntimeEnvironmentStatus.
func (in *CustomRuntimeEnvironmentStatus) DeepCopy() *CustomRuntimeEnvironmentStatus {
	if in == nil {
		return nil
	}
	out := new(CustomRuntimeEnvironmentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalServiceSpec) DeepCopyInto(out *ExternalServiceSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalServiceSpec.
func (in *ExternalServiceSpec) DeepCopy() *ExternalServiceSpec {
	if in == nil {
		return nil
	}
	out := new(ExternalServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImagePullSecret) DeepCopyInto(out *ImagePullSecret) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImagePullSecret.
func (in *ImagePullSecret) DeepCopy() *ImagePullSecret {
	if in == nil {
		return nil
	}
	out := new(ImagePullSecret)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IngressSpec) DeepCopyInto(out *IngressSpec) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IngressSpec.
func (in *IngressSpec) DeepCopy() *IngressSpec {
	if in == nil {
		return nil
	}
	out := new(IngressSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Meteor) DeepCopyInto(out *Meteor) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Meteor.
func (in *Meteor) DeepCopy() *Meteor {
	if in == nil {
		return nil
	}
	out := new(Meteor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Meteor) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteorConfig) DeepCopyInto(out *MeteorConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteorConfig.
func (in *MeteorConfig) DeepCopy() *MeteorConfig {
	if in == nil {
		return nil
	}
	out := new(MeteorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeteorConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteorConfigSpec) DeepCopyInto(out *MeteorConfigSpec) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ControllerManagerConfigurationSpec.DeepCopyInto(&out.ControllerManagerConfigurationSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteorConfigSpec.
func (in *MeteorConfigSpec) DeepCopy() *MeteorConfigSpec {
	if in == nil {
		return nil
	}
	out := new(MeteorConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteorList) DeepCopyInto(out *MeteorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Meteor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteorList.
func (in *MeteorList) DeepCopy() *MeteorList {
	if in == nil {
		return nil
	}
	out := new(MeteorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeteorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteorSpec) DeepCopyInto(out *MeteorSpec) {
	*out = *in
	if in.Pipelines != nil {
		in, out := &in.Pipelines, &out.Pipelines
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteorSpec.
func (in *MeteorSpec) DeepCopy() *MeteorSpec {
	if in == nil {
		return nil
	}
	out := new(MeteorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteorStatus) DeepCopyInto(out *MeteorStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Pipelines != nil {
		in, out := &in.Pipelines, &out.Pipelines
		*out = make([]PipelineResult, len(*in))
		copy(*out, *in)
	}
	in.ExpirationTimestamp.DeepCopyInto(&out.ExpirationTimestamp)
	if in.Comas != nil {
		in, out := &in.Comas, &out.Comas
		*out = make([]NamespacedOwnerReference, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Stage.DeepCopyInto(&out.Stage)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteorStatus.
func (in *MeteorStatus) DeepCopy() *MeteorStatus {
	if in == nil {
		return nil
	}
	out := new(MeteorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespacedOwnerReference) DeepCopyInto(out *NamespacedOwnerReference) {
	*out = *in
	in.OwnerReference.DeepCopyInto(&out.OwnerReference)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespacedOwnerReference.
func (in *NamespacedOwnerReference) DeepCopy() *NamespacedOwnerReference {
	if in == nil {
		return nil
	}
	out := new(NamespacedOwnerReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineResult) DeepCopyInto(out *PipelineResult) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineResult.
func (in *PipelineResult) DeepCopy() *PipelineResult {
	if in == nil {
		return nil
	}
	out := new(PipelineResult)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Shower) DeepCopyInto(out *Shower) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Shower.
func (in *Shower) DeepCopy() *Shower {
	if in == nil {
		return nil
	}
	out := new(Shower)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Shower) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ShowerList) DeepCopyInto(out *ShowerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Shower, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ShowerList.
func (in *ShowerList) DeepCopy() *ShowerList {
	if in == nil {
		return nil
	}
	out := new(ShowerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ShowerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ShowerSpec) DeepCopyInto(out *ShowerSpec) {
	*out = *in
	in.Ingress.DeepCopyInto(&out.Ingress)
	in.Workspace.DeepCopyInto(&out.Workspace)
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ExternalServices != nil {
		in, out := &in.ExternalServices, &out.ExternalServices
		*out = make([]ExternalServiceSpec, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ShowerSpec.
func (in *ShowerSpec) DeepCopy() *ShowerSpec {
	if in == nil {
		return nil
	}
	out := new(ShowerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ShowerStatus) DeepCopyInto(out *ShowerStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ShowerStatus.
func (in *ShowerStatus) DeepCopy() *ShowerStatus {
	if in == nil {
		return nil
	}
	out := new(ShowerStatus)
	in.DeepCopyInto(out)
	return out
}
