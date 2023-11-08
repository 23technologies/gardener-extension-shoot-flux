package validation_test

import (
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"

	. "github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis/flux/v1alpha1"
	. "github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis/flux/v1alpha1/validation"
)

var _ = Describe("CloudProfileConfig validation", func() {
	Describe("#ValidateCloudProfileConfig", func() {
		var (
			rootFldPath *field.Path
			fluxConfig  *FluxConfig
			shoot       *gardencorev1beta1.Shoot
		)

		BeforeEach(func() {
			rootFldPath = field.NewPath("root")

			fluxConfig = &FluxConfig{
				Source: Source{
					Template: sourcev1.GitRepository{
						Spec: sourcev1.GitRepositorySpec{
							Reference: &sourcev1.GitRepositoryRef{
								Branch: "main",
							},
							URL: "https://github.com/fluxcd/flux2-kustomize-helm-example",
						},
					},
				},
				Kustomization: Kustomization{
					Template: kustomizev1.Kustomization{
						Spec: kustomizev1.KustomizationSpec{
							Path: "clusters/production/flux-system",
						},
					},
				},
			}

			shoot = &gardencorev1beta1.Shoot{
				// TODO
			}
		})

		It("should allow basic valid object", func() {
			Expect(ValidateFluxConfig(fluxConfig, shoot, rootFldPath)).To(BeEmpty())
		})

		Describe("FluxInstallation validation", func() {
			BeforeEach(func() {
				fluxConfig.Flux = &FluxInstallation{}
			})

			It("should allow valid namespace names", func() {
				fluxConfig.Flux.Namespace = ptr.To("my-flux-system")

				Expect(ValidateFluxConfig(fluxConfig, shoot, rootFldPath)).To(BeEmpty())
			})

			It("should deny invalid namespace names", func() {
				fluxConfig.Flux.Namespace = ptr.To("this is definitely not a valid namespace")

				Expect(
					ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
				).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("root.flux.namespace"),
				}))))
			})
		})

		Describe("Source validation", func() {
			Describe("TypeMeta validation", func() {
				It("should allow using supported apiVersion and kind", func() {
					fluxConfig.Source.Template.APIVersion = "source.toolkit.fluxcd.io/v1"
					fluxConfig.Source.Template.Kind = "GitRepository"

					Expect(ValidateFluxConfig(fluxConfig, shoot, rootFldPath)).To(BeEmpty())
				})

				It("should allow omitting apiVersion and kind", func() {
					fluxConfig.Source.Template.APIVersion = ""
					fluxConfig.Source.Template.Kind = ""

					Expect(ValidateFluxConfig(fluxConfig, shoot, rootFldPath)).To(BeEmpty())
				})

				It("should deny using unsupported apiVersion", func() {
					fluxConfig.Source.Template.APIVersion = "source.toolkit.fluxcd.io/v2"
					fluxConfig.Source.Template.Kind = "GitRepository"

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeNotSupported),
							"Field": Equal("root.source.template.apiVersion"),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeNotSupported),
							"Field": Equal("root.source.template.kind"),
						})),
					))
				})

				It("should deny using unsupported kind", func() {
					fluxConfig.Source.Template.APIVersion = "source.toolkit.fluxcd.io/v1"
					fluxConfig.Source.Template.Kind = "Bucket"

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeNotSupported),
							"Field": Equal("root.source.template.apiVersion"),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeNotSupported),
							"Field": Equal("root.source.template.kind"),
						})),
					))
				})
			})

			Describe("Reference validation", func() {
				It("should forbid omitting reference", func() {
					fluxConfig.Source.Template.Spec.Reference = nil

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("root.source.template.spec.ref"),
					}))))
				})

				It("should forbid specifying empty reference", func() {
					fluxConfig.Source.Template.Spec.Reference = &sourcev1.GitRepositoryRef{}

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("root.source.template.spec.ref"),
					}))))
				})

				It("should allow setting any reference", func() {
					test := func(mutate func(ref *sourcev1.GitRepositoryRef)) {
						fluxConfig.Source.Template.Spec.Reference = &sourcev1.GitRepositoryRef{}
						mutate(fluxConfig.Source.Template.Spec.Reference)

						ExpectWithOffset(1, ValidateFluxConfig(fluxConfig, shoot, rootFldPath)).To(BeEmpty())
					}

					test(func(ref *sourcev1.GitRepositoryRef) { ref.Branch = "develop" })
					test(func(ref *sourcev1.GitRepositoryRef) { ref.Tag = "latest" })
					test(func(ref *sourcev1.GitRepositoryRef) { ref.SemVer = "v1.0.0" })
					test(func(ref *sourcev1.GitRepositoryRef) { ref.Name = "refs/tags/v1.0.0" })
					test(func(ref *sourcev1.GitRepositoryRef) { ref.Commit = "9c36b9c4bb6438104a703cb983aa8c62ff5e7c4c" })
				})
			})

			Describe("URL validation", func() {
				It("should forbid omitting URL", func() {
					fluxConfig.Source.Template.Spec.URL = ""

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("root.source.template.spec.url"),
					}))))
				})
			})

			Describe("Secret validation", func() {
				It("should allow omitting both secretRef and secretResourceName", func() {
					fluxConfig.Source.Template.Spec.SecretRef = nil
					fluxConfig.Source.SecretResourceName = nil

					Expect(ValidateFluxConfig(fluxConfig, shoot, rootFldPath)).To(BeEmpty())
				})

				It("should allow specifying both secretRef and secretResourceName", func() {
					fluxConfig.Source.Template.Spec.SecretRef = &meta.LocalObjectReference{
						Name: "flux-secret",
					}
					fluxConfig.Source.SecretResourceName = ptr.To("my-flux-secret")
					shoot.Spec.Resources = []gardencorev1beta1.NamedResourceReference{{
						Name: "my-flux-secret",
					}}

					Expect(ValidateFluxConfig(fluxConfig, shoot, rootFldPath)).To(BeEmpty())
				})

				It("should deny specifying a secretResourceName without a matching resource", func() {
					fluxConfig.Source.Template.Spec.SecretRef = &meta.LocalObjectReference{
						Name: "flux-secret",
					}
					fluxConfig.Source.SecretResourceName = ptr.To("my-flux-secret")
					shoot.Spec.Resources = []gardencorev1beta1.NamedResourceReference{{
						Name: "my-other-secret",
					}}

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("root.source.secretResourceName"),
					}))))
				})

				It("should deny omitting secretRef if secretResourceName is set", func() {
					fluxConfig.Source.Template.Spec.SecretRef = nil
					fluxConfig.Source.SecretResourceName = ptr.To("my-flux-secret")
					shoot.Spec.Resources = []gardencorev1beta1.NamedResourceReference{{
						Name: "my-flux-secret",
					}}

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("root.source.template.spec.secretRef"),
					}))))
				})

				It("should deny omitting secretResourceName if secretRef is set", func() {
					fluxConfig.Source.Template.Spec.SecretRef = &meta.LocalObjectReference{
						Name: "flux-secret",
					}
					fluxConfig.Source.SecretResourceName = nil

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("root.source.secretResourceName"),
					}))))
				})
			})
		})

		Describe("Kustomization validation", func() {
			Describe("TypeMeta validation", func() {
				It("should allow using supported apiVersion and kind", func() {
					fluxConfig.Kustomization.Template.APIVersion = "kustomize.toolkit.fluxcd.io/v1"
					fluxConfig.Kustomization.Template.Kind = "Kustomization"

					Expect(ValidateFluxConfig(fluxConfig, shoot, rootFldPath)).To(BeEmpty())
				})

				It("should allow omitting apiVersion and kind", func() {
					fluxConfig.Kustomization.Template.APIVersion = ""
					fluxConfig.Kustomization.Template.Kind = ""

					Expect(ValidateFluxConfig(fluxConfig, shoot, rootFldPath)).To(BeEmpty())
				})

				It("should deny using unsupported apiVersion", func() {
					fluxConfig.Kustomization.Template.APIVersion = "kustomize.toolkit.fluxcd.io/v2"
					fluxConfig.Kustomization.Template.Kind = "Kustomization"

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeNotSupported),
							"Field": Equal("root.kustomization.template.apiVersion"),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeNotSupported),
							"Field": Equal("root.kustomization.template.kind"),
						})),
					))
				})

				It("should deny using unsupported kind", func() {
					fluxConfig.Kustomization.Template.APIVersion = "kustomize.toolkit.fluxcd.io/v1"
					fluxConfig.Kustomization.Template.Kind = "HelmRelease"

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeNotSupported),
							"Field": Equal("root.kustomization.template.apiVersion"),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeNotSupported),
							"Field": Equal("root.kustomization.template.kind"),
						})),
					))
				})
			})

			Describe("Path validation", func() {
				It("should deny omitting path", func() {
					fluxConfig.Kustomization.Template.Spec.Path = ""

					Expect(
						ValidateFluxConfig(fluxConfig, shoot, rootFldPath),
					).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("root.kustomization.template.spec.path"),
					}))))
				})
			})
		})
	})
})
