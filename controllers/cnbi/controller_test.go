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

package cnbi

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	meteorv1alpha1 "github.com/thoth-station/meteor-operator/api/v1alpha1"
)

const (
	timeout  = time.Second * 30
	interval = time.Millisecond * 750
)

var _ = Describe("CustomRuntimeEnvironment controller", func() {
	uni8py38 := meteorv1alpha1.CustomRuntimeEnvironmentRuntimeSpec{
		PythonVersion: "3.8",
		OSName:        "ubi",
		OSVersion:     "8",
	}

	Context("when a CustomRuntimeEnvironment object is created with a RuntimeEnvironment and a PackageList", func() {
		packages := []string{"numpy", "pandas", "scikit-learn"}

		It("should be in Phase 'Running'", func() {
			By("creating a CustomRuntimeEnvironment object")
			build := meteorv1alpha1.BuildTypeSpec{
				BuildType: meteorv1alpha1.PackageList,
			}
			cnbi := &meteorv1alpha1.CustomRuntimeEnvironment{
				TypeMeta:   metav1.TypeMeta{APIVersion: "meteor.zone/v1alpha1", Kind: "CustomRuntimeEnvironment"},
				ObjectMeta: metav1.ObjectMeta{Name: "test-1", Namespace: "default"},
				Spec: meteorv1alpha1.CustomRuntimeEnvironmentSpec{
					RuntimeEnvironment: uni8py38,
					PackageVersions:    packages,
					BuildTypeSpec:      build,
				},
				Status: meteorv1alpha1.CustomRuntimeEnvironmentStatus{},
			}
			Expect(k8sClient.Create(context.Background(), cnbi)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: "test-1", Namespace: "default"}

			Eventually(func(gg Gomega) {
				gg.Consistently(func(g Gomega) {
					createdCRE := &meteorv1alpha1.CustomRuntimeEnvironment{}
					err := k8sClient.Get(ctx, lookupKey, createdCRE)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(createdCRE.Status.Conditions).ToNot(BeEmpty())
					g.Expect(createdCRE.Status.Phase).To(Equal(meteorv1alpha1.PhaseRunning))
				}, "8s", "500ms").Should(Succeed())
			}, timeout, interval).Should(Succeed())

		})
	})
})
