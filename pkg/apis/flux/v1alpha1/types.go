package v1alpha1

import (
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FluxConfig specifies how to bootstrap Flux on the shoot cluster.
// When both "Source" and "Kustomization" are provided they are also installed in the shoot.
// Otherwise, only Flux itself is installed with no Objects to reconcile.
type FluxConfig struct {
	metav1.TypeMeta `json:",inline"`
	// Flux configures the Flux installation in the Shoot cluster.
	// +optional
	Flux *FluxInstallation `json:"flux,omitempty"`
	// Source configures how to bootstrap a Flux source object.
	// If provided, a "Kustomization" must also be provided.
	// +optional
	Source *Source `json:"source,omitempty"`
	// Kustomization configures how to bootstrap a Flux Kustomization object.
	// If provided, "Source" must also be provided.
	// +optional
	Kustomization *Kustomization `json:"kustomization,omitempty"`
}

// FluxInstallation configures the Flux installation in the Shoot cluster.
type FluxInstallation struct {
	// renovate:flux-version
	// renovate updates the doc string. See renovate config for more details

	// Version specifies the Flux version that should be installed.
	// Defaults to "v2.3.0".
	// +optional
	Version *string `json:"version,omitempty"`
	// Registry specifies the container registry where the Flux controller images are pulled from.
	// Defaults to "ghcr.io/fluxcd".
	// +optional
	Registry *string `json:"registry,omitempty"`
	// Namespace specifes the namespace where Flux should be installed.
	// Defaults to "flux-system".
	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

// Source configures how to bootstrap a Flux source object.
type Source struct {
	// Template is a partial GitRepository object in API version source.toolkit.fluxcd.io/v1.
	// Required fields: spec.ref.*, spec.url.
	// The following defaults are applied to omitted field:
	// - metadata.name is defaulted to "flux-system"
	// - metadata.namespace is defaulted to "flux-system"
	// - spec.interval is defaulted to "1m"
	Template sourcev1.GitRepository `json:"template"`
	// SecretResourceName references a resource under Shoot.spec.resources.
	// The secret data from this resource is used to create the GitRepository's credentials secret
	// (GitRepository.spec.secretRef.name) if specified in Template.
	// +optional
	SecretResourceName *string `json:"secretResourceName,omitempty"`
}

// Kustomization configures how to bootstrap a Flux Kustomization object.
type Kustomization struct {
	// Template is a partial Kustomization object in API version kustomize.toolkit.fluxcd.io/v1.
	// Required fields: spec.path.
	// The following defaults are applied to omitted field:
	// - metadata.name is defaulted to "flux-system"
	// - metadata.namespace is defaulted to "flux-system"
	// - spec.interval is defaulted to "1m"
	Template kustomizev1.Kustomization `json:"template"`
}
