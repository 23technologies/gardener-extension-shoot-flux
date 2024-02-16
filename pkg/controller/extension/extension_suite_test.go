package extension

import (
	"context"
	"testing"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	ctx context.Context
	log logr.Logger
)

const (
	poll    = 100 * time.Millisecond
	timeout = 1 * time.Second
)

func TestExtension(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Extension controller Suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()
	log = GinkgoLogr
	SetDefaultEventuallyTimeout(3 * time.Second)
	SetDefaultEventuallyPollingInterval(poll)
})

// testAsync runs f() in a goroutine and returns a channel to wait on. E.g.
//
//	done := testAsync(func(){ Expect(true).To(BeTrue()) })
//	Eventually(done).Should(BeClosed())
//
// This is used here, since the functions under test are blocking until resources become ready.
func testAsync(f func()) chan struct{} {
	done := make(chan struct{})
	go func() {
		defer GinkgoRecover()
		f()
		close(done)
	}()
	return done
}

func newSeedClient() client.Client {
	GinkgoHelper()
	scheme := runtime.NewScheme()
	Expect((&runtime.SchemeBuilder{
		extensionsv1alpha1.AddToScheme,
		clientgoscheme.AddToScheme,
	}).AddToScheme(scheme)).To(Succeed())
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&extensionsv1alpha1.Extension{}).
		Build()
}

func newShootClient() client.Client {
	GinkgoHelper()
	scheme := runtime.NewScheme()
	// kubernetes.NewApplier requires a RESTMapper in order to determine if
	// a resource is namespace-scoped or not. The fake client apparently
	// does not / cannot know this, so CRDs would be considered namespaced
	// (as resources are namespaced by default). So we have to manually
	// inject the information
	mapper := meta.NewDefaultRESTMapper(scheme.PreferredVersionAllGroups())
	mapper.Add(apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"), meta.RESTScopeRoot)
	Expect((&runtime.SchemeBuilder{
		apiextensionsv1.AddToScheme,
		sourcev1.AddToScheme,
		kustomizev1.AddToScheme,
		clientgoscheme.AddToScheme,
	}).AddToScheme(scheme)).To(Succeed())
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRESTMapper(mapper).
		WithStatusSubresource(
			&kustomizev1.Kustomization{},
			&sourcev1.GitRepository{},
		).
		Build()
}
