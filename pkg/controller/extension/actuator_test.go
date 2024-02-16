package extension

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/fluxcd/flux2/v2/pkg/manifestgen"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis/flux/v1alpha1"
)

var _ = Describe("InstallFlux", func() {
	var (
		tmpDir      string
		shootClient client.Client
		config      *v1alpha1.FluxInstallation
	)
	BeforeEach(func() {
		tmpDir = setupManifests()
		shootClient = newShootClient()
		config = &v1alpha1.FluxInstallation{
			Version:   ptr.To("v2.1.3"),
			Registry:  ptr.To("reg.example.com"),
			Namespace: ptr.To("gotk-system"),
		}
	})
	It("succesfully apply and wait for readiness", func() {
		done := testAsync(func() {
			Expect(
				installFlux(ctx, log, shootClient, config, tmpDir, poll, timeout),
			).To(Succeed())
		})
		Eventually(fakeFluxReady(ctx, shootClient, *config.Namespace)).Should(Succeed())

		Eventually(done).Should(BeClosed())
	})
	It("should fail if the resources do not get ready", func() {
		done := testAsync(func() {
			Expect(
				installFlux(ctx, log, shootClient, config, tmpDir, poll, timeout),
			).To(MatchError(ContainSubstring("error waiting for Flux installation to get ready")))
		})

		Eventually(done).Should(BeClosed())
	})
})

var _ = Describe("BootstrapSource", func() {
	var (
		shootClient client.Client
		seedClient  client.Client
		config      *v1alpha1.Source

		extNS = "ext-ns"
	)
	BeforeEach(func() {
		shootClient = newShootClient()
		seedClient = newSeedClient()
		config = &v1alpha1.Source{
			Template: sourcev1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gitrepo",
					Namespace: "custom-namespace",
				},
				Spec: sourcev1.GitRepositorySpec{
					URL: "http://example.com",
				},
			},
		}
	})
	It("should succesfully apply and wait for readiness", func() {
		done := testAsync(func() {
			Expect(
				bootstrapSource(ctx, log, seedClient, shootClient, extNS, nil, config, poll, timeout),
			).To(Succeed())
		})
		repo := config.Template.DeepCopy()
		Eventually(fakeFluxResourceReady(ctx, shootClient, repo)).Should(Succeed())
		Eventually(done).Should(BeClosed())

		createdRepo := &sourcev1.GitRepository{}
		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(repo), createdRepo))
		Expect(createdRepo.Spec.URL).To(Equal("http://example.com"))

		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: config.Template.Namespace}}
		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(ns), ns)).Should(Succeed())
	})
	It("should fail if the resources do not get ready", func() {
		Eventually(testAsync(func() {
			Expect(
				bootstrapSource(ctx, log, seedClient, shootClient, extNS, nil, config, poll, timeout),
			).To(MatchError(ContainSubstring("error waiting for GitRepository to get ready")))
		})).Should(BeClosed())
	})

	Context("with resource secret", func() {
		var (
			secret *corev1.Secret
			ref    []gardencorev1beta1.NamedResourceReference

			refSecretName    = "referenced-secret"
			targetSecretName = "target-secret"
			resourceName     = "the-resource"
		)
		BeforeEach(func() {
			config.Template.Spec.SecretRef = &fluxmeta.LocalObjectReference{
				Name: targetSecretName,
			}
			config.SecretResourceName = &resourceName

			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      v1beta1constants.ReferencedResourcesPrefix + refSecretName,
					Namespace: extNS,
				},
				Data: map[string][]byte{
					"foo": []byte("bar"),
				},
			}
			Expect(seedClient.Create(ctx, secret)).To(Succeed())

			ref = []gardencorev1beta1.NamedResourceReference{
				{
					Name: resourceName,
					ResourceRef: autoscalingv1.CrossVersionObjectReference{
						Name: refSecretName,
					},
				},
			}
		})
		It("should create a referenced resource Secret", func() {
			done := testAsync(func() {
				Expect(
					bootstrapSource(ctx, log, seedClient, shootClient, extNS, ref, config, poll, timeout),
				).To(Succeed())
			})
			repo := config.Template.DeepCopy()
			Eventually(fakeFluxResourceReady(ctx, shootClient, repo)).Should(Succeed())
			Eventually(done).Should(BeClosed())

			createdSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Name:      targetSecretName,
				Namespace: repo.Namespace,
			}}
			Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(createdSecret), createdSecret)).To(Succeed())
			Expect(createdSecret.Data).To(HaveKeyWithValue("foo", []byte("bar")))
		})
	})
})

var _ = Describe("BootstrapKustomization", func() {
	var (
		shootClient client.Client
		config      *v1alpha1.Kustomization
	)
	BeforeEach(func() {
		shootClient = newShootClient()
		config = &v1alpha1.Kustomization{
			Template: kustomizev1.Kustomization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kustomization",
					Namespace: "custom-namespace",
				},
				Spec: kustomizev1.KustomizationSpec{
					Path: "/some/path",
				},
			},
		}
	})
	It("should succesfully apply and wait for readiness", func() {
		done := testAsync(func() {
			Expect(bootstrapKustomization(ctx, log, shootClient, config, poll, timeout)).To(Succeed())
		})
		ks := config.Template.DeepCopy()
		Eventually(fakeFluxResourceReady(ctx, shootClient, ks)).Should(Succeed())
		Eventually(done).Should(BeClosed())

		createdKS := &kustomizev1.Kustomization{}
		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(ks), createdKS))
		Expect(createdKS.Spec.Path).To(Equal("/some/path"))

		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: config.Template.Namespace}}
		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(ns), ns)).Should(Succeed())
	})
	It("should handle if the namespace already exists", func() {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: config.Template.Namespace}}
		Expect(shootClient.Create(ctx, ns)).To(Succeed())

		done := testAsync(func() {
			Expect(bootstrapKustomization(ctx, log, shootClient, config, poll, timeout)).To(Succeed())
		})
		ks := config.Template.DeepCopy()
		Eventually(fakeFluxResourceReady(ctx, shootClient, ks)).Should(Succeed())
		Eventually(done).Should(BeClosed())
	})
	It("should fail if the resources do not get ready", func() {
		Eventually(testAsync(func() {
			Expect(
				bootstrapKustomization(ctx, log, shootClient, config, poll, timeout),
			).To(MatchError(ContainSubstring("error waiting for Kustomization to get ready")))
		})).Should(BeClosed())
	})
})

var _ = Describe("Bootstrapped Condition", func() {
	It("should set and detect a bootstrapped condition", func() {
		seedClient := newSeedClient()
		ext := &extensionsv1alpha1.Extension{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		}
		Expect(seedClient.Create(ctx, ext)).To(Succeed())

		By("being initially false")
		Expect(IsFluxBootstrapped(ext)).To(BeFalse())

		By("setting the bootstrapped condition")
		Expect(SetFluxBootstrapped(ctx, seedClient, ext)).To(Succeed())

		By("reading the condition")
		Expect(seedClient.Get(ctx, client.ObjectKeyFromObject(ext), ext)).To(Succeed())
		Expect(IsFluxBootstrapped(ext)).To(BeTrue())
	})
})

var _ = Describe("GenerateInstallManifest", func() {
	It("should contain the provided options", func() {
		dir := setupManifests()
		out, err := GenerateInstallManifest(&v1alpha1.FluxInstallation{
			Version:   ptr.To("v2.0.0"),
			Registry:  ptr.To("registry.example.com"),
			Namespace: ptr.To("a-namespace"),
		}, dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(And(
			ContainSubstring("v2.0.0"),
			ContainSubstring("registry.example.com"),
			ContainSubstring("a-namespace"),
		))
	})
})

func fakeFluxResourceReady(ctx context.Context, c client.Client, obj fluxmeta.ObjectWithConditionsSetter) func() error {
	return func() error {
		cObj := obj.(client.Object)
		if err := c.Get(ctx, client.ObjectKeyFromObject(cObj), cObj); err != nil {
			return err
		}
		obj.SetConditions([]metav1.Condition{{
			Type:   fluxmeta.ReadyCondition,
			Status: metav1.ConditionTrue,
		}})
		return c.Status().Update(ctx, cObj)
	}
}

func fakeFluxReady(ctx context.Context, c client.Client, namespace string) func() error {
	return func() error {
		gitRepoCRD := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{Name: "gitrepositories." + sourcev1.GroupVersion.Group},
		}
		if err := c.Get(ctx, client.ObjectKeyFromObject(gitRepoCRD), gitRepoCRD); err != nil {
			return err
		}
		gitRepoCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
			{
				Type:   apiextensionsv1.NamesAccepted,
				Status: apiextensionsv1.ConditionTrue,
			},
			{
				Type:   apiextensionsv1.Established,
				Status: apiextensionsv1.ConditionTrue,
			},
		}
		if err := c.Status().Update(ctx, gitRepoCRD); err != nil {
			return err
		}

		sourceController := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "source-controller",
				Namespace: namespace,
			},
		}
		if err := c.Get(ctx, client.ObjectKeyFromObject(sourceController), sourceController); err != nil {
			return err
		}
		sourceController.Status.ObservedGeneration = sourceController.Generation
		sourceController.Status.Conditions = []appsv1.DeploymentCondition{
			{
				Type:   appsv1.DeploymentAvailable,
				Status: corev1.ConditionTrue,
			},
		}
		if err := c.Status().Update(ctx, sourceController); err != nil {
			return err
		}
		return nil
	}
}

// setupManifests copies the local flux manifests to a tmp directory to use for
// tests. This is necessary because flux writes into that directory and we want
// to avoid test pollution.
func setupManifests() string {
	tmpDir, err := manifestgen.MkdirTempAbs("", "gardener-extension-shoot-flux")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() {
		os.RemoveAll(tmpDir)
	})
	srcDir := "./testdata/fluxmanifests"
	files, err := os.ReadDir(srcDir)
	Expect(err).NotTo(HaveOccurred())
	for _, f := range files {
		Expect(copyFile(
			filepath.Join(srcDir, f.Name()),
			filepath.Join(tmpDir, f.Name()),
		)).To(Succeed())
	}
	return tmpDir
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}
