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

const (
	// RequredSecretMissing indicates that the secret required for authentication to the container image registry is missing
	RequiredSecretMissing ConditionType = "RequiredSecretMissing"

	// PipelineRunCreated indicates that the Tekton pipeline run was created
	PipelineRunCreated ConditionType = "PipelineRunCreated"

	// ErrorPipelineRunCreate indicates that the Tekton pipeline run creation failed
	ErrorPipelineRunCreate ConditionType = "ErrorPipelineRunCreate"

	// ValidatingImportedImage indicates that the imported image is being validated by a Tekton PipelineRun's Step
	ValidatingImportedImage ConditionType = "ValidatingImportedImage"

	// PreparingImageBuild indicates that the image build is being prepared by a Tekton PipelineRun's Step
	PreparingImageBuild ConditionType = "PreparingImageBuild"

	// ErrorPreparingImageBuild indicates that the image build preparation failed
	ErrorPreparingImageBuild ConditionType = "ErrorPreparingImageBuild"

	// ErrorResolvingDependencies indicates that the dependency resolution failed during preparation of the image build
	ErrorResolvingDependencies ConditionType = "ErrorResolvingDependencies"

	// BuildingImage indicates that the image is being built by a Tekton PipelineRun
	BuildingImage ConditionType = "BuildingImage"

	// ErrorBuildingImage indicates that the image build failed
	ErrorBuildingImage ConditionType = "ErrorBuildingImage"
)
