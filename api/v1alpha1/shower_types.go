/*
Copyright 2021.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ShowerSpec defines the desired state of Shower
type ShowerSpec struct {
	// Shower UI replicas count
	//+kubebuilder:default=1
	Replicas int32 `json:"replicas"`
	// An optional custom host for Route object
	//+optional
	Ingress IngressSpec `json:"ingress,omitempty"`
	// Workspace PVC setting, defauilts to ReadWriteOnce 500Mi
	//+optional
	Workspace corev1.PersistentVolumeClaimSpec `json:"workspace,omitempty"`
	// Optional shower image. By default the same version as operator is used from quay.io/aicoe/meteor-shower
	//+optional
	Image string `json:"image,omitempty"`
	// Environment variables configuration passed to Shower Deployment spec
	//+optional
	Env []corev1.EnvVar `json:"showerEnv,omitempty"`
	// External services dependencies which can be used by individual pipelines as configurable intergrations e.g. ODH Jupyterhub namespace
	//+optional
	ExternalServices []ExternalServiceSpec `json:"externalServices,omitempty"`
	// Custom host for persistent meteors.
	//+optional
	PersistentMeteorsHost string `json:"persistentMeteorHost,omitempty"`
}

// ShowerStatus defines the observed state of Shower
type ShowerStatus struct {
	// Current condition of the Shower.
	//+optional
	Phase string `json:"phase,omitempty"`
	// Current service state of Meteor.
	//+optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// Most recent observed generation of Shower. Sanity check.
	//+optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Optional shower image. By default the same version as operator is used from quay.io/aicoe/meteor-shower
	//+optional
	Image string `json:"image,omitempty"`
}

// ExternalServiceSpec defines external integration point wich can be used by pipelines submitted by Meteor
type ExternalServiceSpec struct {
	Name string `json:"name"`
	//+optional
	Namespace string `json:"namespace,omitempty"`
	//+optional
	Url string `json:"url,omitempty"`
}

// IngressSpec configures Route resource exposed by the Shower deployment
type IngressSpec struct {
	//+optional
	Annotations map[string]string `json:"annotations,omitempty"`
	//+optional
	Labels map[string]string `json:"labels,omitempty"`
	//+optional
	Host string `json:"host,omitempty"`
	//+optional
	Path string `json:"path,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+operator-sdk:csv:customresourcedefinitions:resources={{Role,v1},{RoleBinding,v1},{Deployment,v1},{Service,v1},{Route,v1},{ServiceAccount,v1},{ServiceMonitor,v1}}

// Shower represents a Shower UI and runtime configuration associated with Meteors produced from this instance.
type Shower struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShowerSpec   `json:"spec,omitempty"`
	Status ShowerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ShowerList contains a list of Shower
type ShowerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Shower `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Shower{}, &ShowerList{})
}

// Aggregate phase from conditions
func (m *Shower) AggregatePhase() string {
	if len(m.Status.Conditions) == 0 {
		return PhasePending
	}

	for _, c := range m.Status.Conditions {
		switch c.Type {
		case "Deployment":
			if c.Status == metav1.ConditionFalse {
				return PhasePending
			}
		}
	}
	return PhaseOk
}

func (m *Shower) GetReference(isController bool) NamespacedOwnerReference {
	blockOwnerDeletion := true
	return NamespacedOwnerReference{
		OwnerReference: metav1.OwnerReference{
			APIVersion:         m.APIVersion,
			Kind:               m.Kind,
			Name:               m.GetName(),
			UID:                m.GetUID(),
			Controller:         &isController,
			BlockOwnerDeletion: &blockOwnerDeletion,
		},
		Namespace: m.GetNamespace(),
	}
}
