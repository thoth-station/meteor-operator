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

// Phase describes the phase of the CustomRuntimeEnvironment
// +kubebuilder:validation:Enum=Pending;Failed;Running;Succeeded;Unknown
type Phase string

const (
	PhasePending   = Phase("Pending")
	PhaseFailed    = Phase("Failed")
	PhaseRunning   = Phase("Running")
	PhaseBuilding  = Phase("Building")
	PhaseSucceeded = Phase("Succeeded")
	PhaseUnknown   = Phase("Unknown")
)
