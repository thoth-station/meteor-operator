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
	// PipelineRunCreated indicates that the Tekton pipeline run was created
	PipelineRunCreated = "PipelineRunCreated"

	// ErrorPipelineRunCreate indicates that the Tekton pipeline run creation failed
	ErrorPipelineRunCreate = "ErrorPipelineRunCreate"

	// ImportingImage indicates that the image is being imported from a remote registry
	ImportingImage = "ImportingImage"

	// RequredSecretMissing indicates that the secret required for authentication to the container image registry is missing
	RequiredSecretMissing = "RequiredSecretMissing"

	// ValidatingImportedImage indicates that the imported image is being validated by a Tekton PipelineRun's Step
	ValidatingImportedImage = "ValidatingImportedImage"

	// ImageImportReady indicates that the imported image is ready to be used
	ImageImportReady = "ImageImportReady"

	// ImageImportInvalid indicates that the imported image is invalid
	ImageImportInvalid = "ImageImportInvalid"

	// ErrorResolvingDependencies indicates that the dependency resolution failed during preparation of the image build
	ErrorResolvingDependencies = "ErrorResolvingDependencies"

	// BuildingImage indicates that the image is being built by a Tekton PipelineRun
	BuildingImage = "BuildingImage"

	// PackageListBuildCompleted indicates that the package list build completed
	PackageListBuildCompleted = "PackageListBuildCompleted"

	// ErrorBuildingImage indicates that the image build failed
	ErrorBuildingImage = "ErrorBuildingImage"

	// GenericPipelineError indicates that the pipeline failed with an error
	GenericPipelineError = "GenericPipelineError"

	// PipelineRunCompleted indicates that the Tekton pipeline run completed
	PipelineRunCompleted = "PipelineRunCompleted"
)
