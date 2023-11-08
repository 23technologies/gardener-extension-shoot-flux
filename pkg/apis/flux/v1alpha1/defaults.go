package v1alpha1

import (
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_FluxConfig(obj *FluxConfig) {
	if obj.Flux == nil {
		obj.Flux = &FluxInstallation{}
	}
}

func SetDefaults_FluxInstallation(obj *FluxInstallation) {
	if obj.Version == nil {
		// TODO: add renovate to upgrade this in lockstep with go modules update
		//  If you do automate this, also automate updating the API doc string ;)
		obj.Version = ptr.To("v2.1.2")
	}

	if obj.Registry == nil {
		obj.Registry = ptr.To("ghcr.io/fluxcd")
	}

	if obj.Namespace == nil {
		obj.Namespace = ptr.To("flux-system")
	}
}

func SetDefaults_Source(obj *Source) {
	SetDefaults_Flux_GitRepository(&obj.Template)

	hasSecretRef := obj.Template.Spec.SecretRef != nil && obj.Template.Spec.SecretRef.Name != ""
	hasSecretResourceName := ptr.Deref(obj.SecretResourceName, "") != ""
	if hasSecretResourceName && !hasSecretRef {
		obj.Template.Spec.SecretRef = &meta.LocalObjectReference{
			Name: "flux-system",
		}
	}
}

func SetDefaults_Kustomization(obj *Kustomization) {
	SetDefaults_Flux_Kustomization(&obj.Template)
}

func SetDefaults_Flux_GitRepository(obj *sourcev1.GitRepository) {
	if obj.Name == "" {
		obj.Name = "flux-system"
	}

	if obj.Namespace == "" {
		obj.Namespace = "flux-system"
	}

	if obj.Spec.Interval.Duration == 0 {
		obj.Spec.Interval = metav1.Duration{Duration: time.Minute}
	}
}

func SetDefaults_Flux_Kustomization(obj *kustomizev1.Kustomization) {
	if obj.Name == "" {
		obj.Name = "flux-system"
	}

	if obj.Namespace == "" {
		obj.Namespace = "flux-system"
	}

	if obj.Spec.Interval.Duration == 0 {
		obj.Spec.Interval = metav1.Duration{Duration: time.Minute}
	}
}
