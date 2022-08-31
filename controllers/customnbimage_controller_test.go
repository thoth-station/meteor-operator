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

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
)

const (
	timeout  = time.Second * 10
	duration = time.Second * 10
	interval = time.Millisecond * 750
)

var _ = Describe("CustomNBImage controller", func() {
	Context("when a CustomNBImage object is created with a RuntimeEnvironment", func() {
		It("should have Status 'Preparing'", func() {
			// TODO implement your test here
			By("creating a CustomNBImage object")
			cnbi := &meteorv1alpha1.CustomNBImage{
				TypeMeta:   metav1.TypeMeta{APIVersion: "meteor.zone/v1alpha1", Kind: "CustomNBImage"},
				ObjectMeta: metav1.ObjectMeta{Name: "custom-nbimage-1", Namespace: "default"},
				Spec: meteorv1alpha1.CustomNBImageSpec{
					RuntimeEnvironment: meteorv1alpha1.CustomNBImageRuntimeSpec{
						PythonVersion: "3.8",
						OSName:        "ubi",
						OSVersion:     "8",
					},
				},
				Status: meteorv1alpha1.CustomNBImageStatus{},
			}
			Expect(k8sClient.Create(context.Background(), cnbi)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: "custom-nbimage-1", Namespace: "default"}
			createdCNBi := &meteorv1alpha1.CustomNBImage{}

			Consistently(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdCNBi)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdCNBi.Status.Phase).Should(Equal(meteorv1alpha1.CNBiPhasePreparing))
		})
	})
	Context("when a CustomNBImage object is created with Import Strategy", func() {
		It("should have Status 'Importing'", func() {
			// TODO implement your test here
			By("creating a CustomNBImage object")
			cnbi := &meteorv1alpha1.CustomNBImage{
				TypeMeta:   metav1.TypeMeta{APIVersion: "meteor.zone/v1alpha1", Kind: "CustomNBImage"},
				ObjectMeta: metav1.ObjectMeta{Name: "custom-nbimage-2", Namespace: "default"},
				Spec: meteorv1alpha1.CustomNBImageSpec{
					Strategy: meteorv1alpha1.CustomNBImageStrategy{
						Type: "import",
						From: "quay.io/thoth-station/s2i-minimal-py38-notebook:v0.2.2",
					},
				},
				Status: meteorv1alpha1.CustomNBImageStatus{},
			}
			Expect(k8sClient.Create(context.Background(), cnbi)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: "custom-nbimage-2", Namespace: "default"}
			createdCNBi := &meteorv1alpha1.CustomNBImage{}

			Consistently(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdCNBi)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdCNBi.Status.Phase).Should(Equal(meteorv1alpha1.CNBiPhaseImporting))
		})
	})
})
