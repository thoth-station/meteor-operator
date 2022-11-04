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
	timeout  = time.Second * 10
	duration = time.Second * 10
	interval = time.Millisecond * 750
)

var _ = Describe("CustomNBImage controller", func() {
	uni8py38 := meteorv1alpha1.CustomNBImageRuntimeSpec{
		PythonVersion: "3.8",
		OSName:        "ubi",
		OSVersion:     "8",
	}

	Context("when a CustomNBImage object is created with a RuntimeEnvironment and a PackageList", func() {
		packages := []string{"numpy", "pandas", "scikit-learn"}

		It("should be in Phase 'Running'", func() {
			By("creating a CustomNBImage object")
			build := meteorv1alpha1.BuildTypeSpec{
				BuildType: meteorv1alpha1.PackageList,
			}
			cnbi := &meteorv1alpha1.CustomNBImage{
				TypeMeta:   metav1.TypeMeta{APIVersion: "meteor.zone/v1alpha1", Kind: "CustomNBImage"},
				ObjectMeta: metav1.ObjectMeta{Name: "test-1", Namespace: "default"},
				Spec: meteorv1alpha1.CustomNBImageSpec{
					RuntimeEnvironment: uni8py38,
					PackageVersions:    packages,
					BuildTypeSpec:      build,
				},
				Status: meteorv1alpha1.CustomNotebookImageStatus{},
			}
			Expect(k8sClient.Create(context.Background(), cnbi)).Should(Succeed())
			time.Sleep(20 * time.Second) // FIXME 👻 smells like a race condition, please increase the timeout for slow clusters

			lookupKey := types.NamespacedName{Name: "test-1", Namespace: "default"}
			createdCNBi := &meteorv1alpha1.CustomNBImage{}

			Consistently(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdCNBi)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdCNBi.Status.Conditions).ShouldNot(BeEmpty())
			Expect(createdCNBi.Status.Phase).Should(Equal(meteorv1alpha1.PhaseRunning))
		})
	})
	/*
		Context("when a CustomNBImage object is created with ImportImage BuildType", func() {
			It("should have Condition 'Importing'", func() {
				By("creating a CustomNBImage object")
				build := meteorv1alpha1.BuildTypeSpec{
					BuildType: meteorv1alpha1.ImportImage,
					FromImage: "quay.io/thoth-station/s2i-custom-notebook:latest",
				}
				cnbi := &meteorv1alpha1.CustomNBImage{
					TypeMeta:   metav1.TypeMeta{APIVersion: "meteor.zone/v1alpha1", Kind: "CustomNBImage"},
					ObjectMeta: metav1.ObjectMeta{Name: "test-2", Namespace: "default"},
					Spec: meteorv1alpha1.CustomNBImageSpec{
						RuntimeEnvironment: meteorv1alpha1.CustomNBImageRuntimeSpec{},
						PackageVersions:    []string{},
						BuildTypeSpec:      build,
					},
					Status: meteorv1alpha1.CustomNotebookImageStatus{},
				}
				Expect(k8sClient.Create(context.Background(), cnbi)).Should(Succeed())

				By("checking the CustomNBImage object has been created on the cluster")
				// lets give the cluster a little time to start reconciling
				time.Sleep(20 * time.Second)

				lookupKey := types.NamespacedName{Name: "test-2", Namespace: "default"}
				createdCNBi := &meteorv1alpha1.CustomNBImage{}

				Eventually(func() bool {
					err := k8sClient.Get(ctx, lookupKey, createdCNBi)
					return err == nil
				}, timeout, interval).Should(BeTrue())

				By("looking if the Controller started reconciling the CustomNBImage object")
				Expect(createdCNBi.Status.Phase).Should(Equal(meteorv1alpha1.CNBiPhaseRunning))
				Expect(createdCNBi.Status.Conditions).ShouldNot(BeEmpty())
				Expect(createdCNBi.Status.Conditions[0].Type).Should(Equal(meteorv1alpha1.ImportingImage))
			})
			It("should have Condition 'RequiredSecretMissing' if the import is from a repo that reqs auth and Secret is not ready", func() {
				By("creating a CustomNBImage object for an import from a private repository")
				cnbi_name := "import-private-repository"

				importFromPrivate := meteorv1alpha1.BuildTypeSpec{
					BuildType: meteorv1alpha1.ImportImage,
					FromImage: "quay.io/thoth-station/s2i-custom-notebook:latest",
					ImagePullSecret: meteorv1alpha1.ImagePullSecret{
						Name: "private-repository-credentials",
					},
				}
				cnbi := &meteorv1alpha1.CustomNBImage{
					TypeMeta:   metav1.TypeMeta{APIVersion: "meteor.zone/v1alpha1", Kind: "cnbi"},
					ObjectMeta: metav1.ObjectMeta{Name: cnbi_name, Namespace: "default"},
					Spec: meteorv1alpha1.CustomNBImageSpec{
						RuntimeEnvironment: meteorv1alpha1.CustomNBImageRuntimeSpec{},
						BuildTypeSpec:      importFromPrivate,
					},
					Status: meteorv1alpha1.CustomNotebookImageStatus{},
				}
				Expect(k8sClient.Create(context.Background(), cnbi)).Should(Succeed())

				By("checking the CustomNBImage object has been created on the cluster")
				// lets give the cluster a little time to start reconciling
				time.Sleep(8 * time.Second)

				lookupKey := types.NamespacedName{Name: cnbi_name, Namespace: "default"}
				createdCNBi := &meteorv1alpha1.CustomNBImage{}

				Eventually(func() bool {
					err := k8sClient.Get(ctx, lookupKey, createdCNBi)
					return err == nil
				}, timeout, interval).Should(BeTrue())

				By("looking if the Controller started reconciling the CustomNBImage object")
				Expect(createdCNBi.Status.Phase).Should(Equal(meteorv1alpha1.CNBiPhaseRunning))
				Expect(createdCNBi.Status.Conditions).ShouldNot(BeEmpty())
				Expect(createdCNBi.Status.Conditions[0].Type).Should(Equal(meteorv1alpha1.RequiredSecretMissing))
			})
		}) */
})
