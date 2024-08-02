package extension

import (
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	. "github.com/gardener/gardener/pkg/utils/test/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fluxv1alpha1 "github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis/flux/v1alpha1"
)

var _ = Describe("ReconcileSecrets", Ordered, func() {
	var (
		shootClient client.Client
		seedClient  client.Client

		config    *fluxv1alpha1.FluxConfig
		resources []gardencorev1beta1.NamedResourceReference
		extNS     = "ext-ns"
	)
	BeforeAll(func() {
		shootClient = newShootClient()
		seedClient = newSeedClient()
		config = &fluxv1alpha1.FluxConfig{
			Flux: &fluxv1alpha1.FluxInstallation{
				Namespace: ptr.To("flux-system"),
			},
			Source: &fluxv1alpha1.Source{
				SecretResourceName: ptr.To("source-secret"),
				Template: sourcev1.GitRepository{
					Spec: sourcev1.GitRepositorySpec{
						SecretRef: &fluxmeta.LocalObjectReference{
							Name: "ssh-target-name",
						},
					},
				},
			},
			AdditionalSecretResources: []fluxv1alpha1.AdditionalResource{{
				Name: "extra-secret",
			}},
		}
		resources = []gardencorev1beta1.NamedResourceReference{
			{
				Name: "source-secret",
				ResourceRef: autoscalingv1.CrossVersionObjectReference{
					Name: "ssh",
					Kind: "Secret",
				},
			},
			{
				Name: "extra-secret",
				ResourceRef: autoscalingv1.CrossVersionObjectReference{
					Name: "extra",
					Kind: "Secret",
				},
			},
		}
		Expect(seedClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ref-ssh",
				Namespace: extNS,
			},
			Data: map[string][]byte{
				"foo": []byte("ssh"),
			},
		})).To(Succeed())
		Expect(seedClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ref-extra",
				Namespace: extNS,
			},
			Data: map[string][]byte{
				"foo": []byte("extra"),
			},
		})).To(Succeed())
	})

	It("should create a referenced Source Secret", func() {
		Expect(
			ReconcileSecrets(ctx, log, seedClient, shootClient, extNS, config, resources),
		).To(Succeed())

		createdSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      "ssh-target-name",
			Namespace: "flux-system",
		}}
		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(createdSecret), createdSecret)).To(Succeed())
		Expect(createdSecret.Data).To(HaveKeyWithValue("foo", []byte("ssh")))
		Expect(createdSecret.Labels).To(HaveKeyWithValue(managedByLabelKey, managedByLabelValue))
	})

	It("should create the additional secrets", func() {
		Expect(
			ReconcileSecrets(ctx, log, seedClient, shootClient, extNS, config, resources),
		).To(Succeed())

		createdSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      "extra",
			Namespace: "flux-system",
		}}
		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(createdSecret), createdSecret)).To(Succeed())
		Expect(createdSecret.Data).To(HaveKeyWithValue("foo", []byte("extra")))
	})

	It("should not change an existing secret", func() {
		createdSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      "extra",
			Namespace: "flux-system",
		}}
		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(createdSecret), createdSecret)).To(Succeed())
		createdSecret.Data["foo"] = []byte("changed")
		Expect(shootClient.Update(ctx, createdSecret)).To(Succeed())

		Expect(
			ReconcileSecrets(ctx, log, seedClient, shootClient, extNS, config, resources),
		).To(Succeed())

		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(createdSecret), createdSecret)).To(Succeed())
		Expect(createdSecret.Data).To(HaveKeyWithValue("foo", []byte("changed")))
	})

	It("should respect the target name and clean up the old secret", func() {
		config.AdditionalSecretResources[0].TargetName = ptr.To("surprise")
		Expect(
			ReconcileSecrets(ctx, log, seedClient, shootClient, extNS, config, resources),
		).To(Succeed())

		createdSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      "surprise",
			Namespace: "flux-system",
		}}
		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(createdSecret), createdSecret)).To(Succeed())
		Expect(createdSecret.Data).To(HaveKeyWithValue("foo", []byte("extra")))

		deletedSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Name:      "extra",
			Namespace: "flux-system",
		}}
		Expect(shootClient.Get(ctx, client.ObjectKeyFromObject(deletedSecret), deletedSecret)).To(BeNotFoundError())
	})
})
