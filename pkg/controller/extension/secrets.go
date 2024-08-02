package extension

import (
	"context"
	"fmt"
	"maps"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	v1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fluxv1alpha1 "github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis/flux/v1alpha1"
)

const (
	managedByLabelKey   = "app.kubernetes.io/managed-by"
	managedByLabelValue = "gardener-extension-" + fluxv1alpha1.ExtensionType
)

// ReconcileSecrets copies all secrets referenced in the extension (additionalSecretResources or
// source.SecretResourceName), and deletes all secrets that are no longer referenced.
// We cannot use gardener resource manager here, because we want to work in the namespace
// "flux-system", which the resource manager is not configured for.
func ReconcileSecrets(
	ctx context.Context,
	log logr.Logger,
	seedClient client.Client,
	shootClient client.Client,
	seedNamespace string,
	config *fluxv1alpha1.FluxConfig,
	resources []gardencorev1beta1.NamedResourceReference,
) error {
	shootNamespace := *config.Flux.Namespace
	secretsToKeep := sets.Set[string]{}

	secretResources := config.AdditionalSecretResources
	if config.Source != nil && config.Source.SecretResourceName != nil {
		secretResources = append(secretResources, fluxv1alpha1.AdditionalResource{
			Name:       *config.Source.SecretResourceName,
			TargetName: ptr.To(config.Source.Template.Spec.SecretRef.Name),
		})
	}
	for _, resource := range secretResources {
		name, err := copySecretToShoot(ctx, log, seedClient, shootClient, seedNamespace, shootNamespace, resources, resource)
		if err != nil {
			return fmt.Errorf("failed to copy secret: %w", err)
		}
		secretsToKeep.Insert(name)
	}

	// cleanup unreferenced secrets
	secretList := &corev1.SecretList{}
	if err := shootClient.List(ctx, secretList,
		client.InNamespace(shootNamespace),
		client.MatchingLabels{managedByLabelKey: managedByLabelValue},
	); err != nil {
		return fmt.Errorf("failed to list managed secrets in shoot: %w", err)
	}
	for _, secret := range secretList.Items {
		if secretsToKeep.Has(secret.Name) {
			continue
		}
		if err := shootClient.Delete(ctx, &secret); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete secret that is no longer referenced: %w", err)
		}
		log.Info("Deleted secret that is no longer referenced by the extension", "secretName", secret.Name)
	}
	return nil
}

func copySecretToShoot(
	ctx context.Context,
	log logr.Logger,
	seedClient client.Client,
	shootClient client.Client,
	seedNamespace string,
	targetNamespace string,
	resources []gardencorev1beta1.NamedResourceReference,
	additionalResource fluxv1alpha1.AdditionalResource,
) (string, error) {
	resource := v1beta1helper.GetResourceByName(resources, additionalResource.Name)
	if resource == nil {
		return "", fmt.Errorf("secret resource name does not match any of the resource names in Shoot.spec.resources[].name")
	}

	seedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v1beta1constants.ReferencedResourcesPrefix + resource.ResourceRef.Name,
			Namespace: seedNamespace,
		},
	}
	if err := seedClient.Get(ctx, client.ObjectKeyFromObject(seedSecret), seedSecret); err != nil {
		return "", fmt.Errorf("error reading referenced secret: %w", err)
	}

	name := resource.ResourceRef.Name
	if additionalResource.TargetName != nil {
		name = *additionalResource.TargetName
	}
	shootSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: targetNamespace,
		},
	}
	err := shootClient.Get(ctx, client.ObjectKeyFromObject(shootSecret), shootSecret)
	if client.IgnoreNotFound(err) != nil {
		return "", err
	}
	if err == nil {
		return shootSecret.Name, nil
	}

	shootSecret.Data = maps.Clone(seedSecret.Data)
	shootSecret.Labels = map[string]string{
		managedByLabelKey: managedByLabelValue,
	}

	if err := shootClient.Create(ctx, shootSecret); client.IgnoreAlreadyExists(err) != nil {
		return "", err
	}
	log.Info("Created secret", "secretName", shootSecret.Name)

	return shootSecret.Name, nil
}
