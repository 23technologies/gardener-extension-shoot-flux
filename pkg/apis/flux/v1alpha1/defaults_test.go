package v1alpha1

import (
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	. "github.com/gardener/gardener/pkg/utils/test/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("FluxConfig defaulting", func() {
	var obj *FluxConfig

	BeforeEach(func() {
		obj = &FluxConfig{
			Source: &Source{
				Template: sourcev1.GitRepository{
					Spec: sourcev1.GitRepositorySpec{
						Reference: &sourcev1.GitRepositoryRef{
							Branch: "main",
						},
						URL: "https://github.com/fluxcd/flux2-kustomize-helm-example",
					},
				},
			},
			Kustomization: &Kustomization{
				Template: kustomizev1.Kustomization{
					Spec: kustomizev1.KustomizationSpec{
						Path: "clusters/production/flux-system",
					},
				},
			},
		}
	})

	It("should not overwrite required fields", func() {
		before := obj.DeepCopy()

		SetObjectDefaults_FluxConfig(obj)

		Expect(obj.Source.Template.Spec.Reference).To(DeepEqual(before.Source.Template.Spec.Reference))
		Expect(obj.Source.Template.Spec.URL).To(DeepEqual(before.Source.Template.Spec.URL))
		Expect(obj.Source.Template.Spec.URL).To(DeepEqual(before.Source.Template.Spec.URL))
		Expect(obj.Kustomization.Template.Spec.Path).To(DeepEqual(before.Kustomization.Template.Spec.Path))
	})

	Describe("FluxInstallation defaulting", func() {
		It("should default all standard fields", func() {
			SetObjectDefaults_FluxConfig(obj)

			Expect(obj.Flux).To(DeepEqual(&FluxInstallation{
				Version:   ptr.To(defaultFluxVersion),
				Registry:  ptr.To("ghcr.io/fluxcd"),
				Namespace: ptr.To("flux-system"),
			}))
		})
	})

	Describe("Source defaulting", func() {
		It("should default all standard fields", func() {
			SetObjectDefaults_FluxConfig(obj)

			Expect(obj.Source.Template.Name).To(Equal("flux-system"))
			Expect(obj.Source.Template.Namespace).To(Equal("flux-system"))
			Expect(obj.Source.Template.Spec.Interval.Duration).To(Equal(time.Minute))
		})

		It("should default secretRef.name to flux-system if secretResourceName is set", func() {
			obj.Source.SecretResourceName = ptr.To("my-flux-secret")

			SetObjectDefaults_FluxConfig(obj)

			Expect(obj.Source.Template.Spec.SecretRef).NotTo(BeNil())
			Expect(obj.Source.Template.Spec.SecretRef.Name).To(Equal("flux-system"))
		})

		It("should not overwrite secretRef.name if secretResourceName is set", func() {
			obj.Source.Template.Spec.SecretRef = &meta.LocalObjectReference{Name: "flux-secret"}
			obj.Source.SecretResourceName = ptr.To("my-flux-secret")

			SetObjectDefaults_FluxConfig(obj)

			Expect(obj.Source.Template.Spec.SecretRef).NotTo(BeNil())
			Expect(obj.Source.Template.Spec.SecretRef.Name).To(Equal("flux-secret"))
		})

		It("should handle if the kustomization is omitted", func() {
			obj.Kustomization = nil
			SetObjectDefaults_FluxConfig(obj)

			Expect(obj.Source.Template.Name).To(Equal("flux-system"))
			Expect(obj.Source.Template.Namespace).To(Equal("flux-system"))
		})
	})

	Describe("Kustomization defaulting", func() {
		It("should default all standard fields", func() {
			SetObjectDefaults_FluxConfig(obj)

			Expect(obj.Kustomization.Template.Name).To(Equal("flux-system"))
			Expect(obj.Kustomization.Template.Namespace).To(Equal("flux-system"))
			Expect(obj.Kustomization.Template.Spec.Interval.Duration).To(Equal(time.Minute))
		})
		It("should handle if the source is omitted", func() {
			obj.Source = nil
			SetObjectDefaults_FluxConfig(obj)
			Expect(obj.Kustomization.Template.Name).To(Equal("flux-system"))
			Expect(obj.Kustomization.Template.Namespace).To(Equal("flux-system"))
		})
	})
})
