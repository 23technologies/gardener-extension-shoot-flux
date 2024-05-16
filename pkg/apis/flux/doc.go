// Package flux is a dummy package to make k8s.io/code-generator work without an internal version.

// +k8s:deepcopy-gen=package
// +groupName=flux.extensions.gardener.cloud

//go:generate ../../../hack/update-codegen.sh

package flux // import "github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis/flux"
