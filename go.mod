module github.com/23technologies/gardener-extension-shoot-flux

go 1.16

require (
	github.com/fluxcd/flux2 v0.29.0
	github.com/fluxcd/kustomize-controller/api v0.24.0
	github.com/fluxcd/pkg/apis/meta v0.12.2
	github.com/fluxcd/source-controller/api v0.24.0
	github.com/gardener/gardener v1.43.1
	github.com/go-logr/logr v1.2.2
	github.com/google/go-github/v44 v44.0.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/spf13/cobra v1.4.0
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/component-base v0.23.5
	sigs.k8s.io/controller-runtime v0.11.2
	sigs.k8s.io/yaml v1.3.0
)

replace (
	// github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/apiserver => k8s.io/apiserver v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.2
	k8s.io/code-generator => k8s.io/code-generator v0.22.2
	k8s.io/component-base => k8s.io/component-base v0.22.2
	k8s.io/helm => k8s.io/helm v2.13.1+incompatible
// sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.10.2
)
