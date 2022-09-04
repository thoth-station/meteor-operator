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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CustomNBImage Webhook", func() {
	build := BuildTypeSpec{
		Type: PackageList,
	}

	Context("when a CustomNBImage object is created", func() {
		It("should pass if all required annotations are present", func() {
			By("creating a CustomNBImage object")
			cnbi := &CustomNBImage{
				TypeMeta:   metav1.TypeMeta{APIVersion: "meteor.zone/v1alpha1", Kind: "CustomNBImage"},
				ObjectMeta: metav1.ObjectMeta{Name: "webhook-1", Namespace: "default"},
				Spec: CustomNBImageSpec{
					BuildTypeSpec: build,
				},
				Status: CustomNBImageStatus{},
			}
			metav1.SetMetaDataAnnotation(&cnbi.ObjectMeta, CNBiNameAnnotationKey, "webhook-1")
			metav1.SetMetaDataAnnotation(&cnbi.ObjectMeta, CNBiDescriptionAnnotationKey, "default")
			metav1.SetMetaDataAnnotation(&cnbi.ObjectMeta, CNBiCreatorAnnotationKey, "ginko+gomega")

			Expect(k8sClient.Create(context.Background(), cnbi)).Should(Succeed())

		})
		It("should fail if annotations are missing completely", func() {
			By("creating an inclomplte CustomNBImage object")
			cnbi := &CustomNBImage{
				TypeMeta:   metav1.TypeMeta{APIVersion: "meteor.zone/v1alpha1", Kind: "CustomNBImage"},
				ObjectMeta: metav1.ObjectMeta{Name: "webhook-2", Namespace: "default"},
				Spec: CustomNBImageSpec{
					BuildTypeSpec: build,
				},
				Status: CustomNBImageStatus{},
			}

			Expect(k8sClient.Create(context.Background(), cnbi)).ShouldNot(Succeed())
		})
		It("should fail if on annotation is missing", func() {
			By("creating an inclomplte CustomNBImage object")
			cnbi := &CustomNBImage{
				TypeMeta:   metav1.TypeMeta{APIVersion: "meteor.zone/v1alpha1", Kind: "CustomNBImage"},
				ObjectMeta: metav1.ObjectMeta{Name: "webhook-3", Namespace: "default"},
				Spec: CustomNBImageSpec{
					BuildTypeSpec: build,
				},
				Status: CustomNBImageStatus{},
			}
			metav1.SetMetaDataAnnotation(&cnbi.ObjectMeta, CNBiNameAnnotationKey, "webhook-3")

			err := k8sClient.Create(context.Background(), cnbi)
			Expect(err).ShouldNot(Succeed())
			Expect(err).Should(MatchError("admission webhook \"vcustomnbimage.kb.io\" denied the request: CustomNBImage.meteor.zone \"webhook-3\" is invalid: [metadata.annotations[opendatahub.io/notebook-image-desc]: Required value: annotation is required, metadata.annotations[opendatahub.io/notebook-image-creator]: Required value: annotation is required]"))
		})

	})
})
