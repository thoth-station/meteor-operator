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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	config "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MeteorConfigSpec defines the desired state of MeteorConfig
type MeteorConfigSpec struct {
	metav1.TypeMeta `json:",inline"`

	// ControllerManagerConfigurationSpec returns the contfigurations for controllers
	config.ControllerManagerConfigurationSpec `json:",inline"`

	// EnableShower is the feature flar/config to enable Shower
	EnableShower bool `json:"enableComa,omitempty"`
}

//+kubebuilder:object:root=true

// MeteorConfig is the Schema for the meteorconfigs API
type MeteorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MeteorConfigSpec `json:"spec,omitempty"`
}

// Complete implements config.ControllerManagerConfiguration
func (*MeteorConfig) Complete() (config.ControllerManagerConfigurationSpec, error) {
	return config.ControllerManagerConfigurationSpec{}, nil
}

func init() {
	SchemeBuilder.Register(&MeteorConfig{})
}
