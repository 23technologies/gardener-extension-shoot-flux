// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package lifecycle contains functions used at the lifecycle controller
package lifecycle

import (
	"context"
	_ "embed"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-github/v44/github"

	"github.com/23technologies/gardener-extension-shoot-flux/pkg/constants"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	gardenclient "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/extensions"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	managedresources "github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/gardener/gardener/pkg/utils/retry"
	"github.com/go-logr/logr"
	"sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/fluxcd/flux2/pkg/manifestgen/sourcesecret"
	kustomizecontrollerv1beta2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcecontrollerv1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
)

const (
	// ActuatorName is the name of the Shoot Flux actuator.
	ActuatorName = constants.ServiceName + "-actuator"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator() extension.Actuator {
	return &actuator{
		logger: log.Log.WithName(ActuatorName),
	}
}

type actuator struct {
	client          client.Client // controller-runtime client for interaction with the seed cluster
	clientGardenlet client.Client // controller-runtime client for interaction with the garden cluster
	config          *rest.Config
	decoder         runtime.Decoder
	logger          logr.Logger // logger
}

// Reconcile the Extension resource.
//
// PARAMETERS
// ctx context.Context               Execution context
// ex  *extensionsv1alpha1.Extension Extension struct
func (a *actuator) Reconcile(ctx context.Context, ex *extensionsv1alpha1.Extension) error {

	// get the shoot and the project namespace
	extensionNamespace := ex.GetNamespace()
	shoot, err := extensions.GetShoot(ctx, a.client, extensionNamespace)
	projectNamespace := shoot.GetNamespace()

	// fetch the configmap holding the per-project configuration for the current flux installation
	fluxConfigMap := corev1.ConfigMap{}
	err = a.clientGardenlet.Get(ctx, kutil.Key(projectNamespace, "flux-config"), &fluxConfigMap)
	if err != nil {
		return err
	}

	fluxVersion, err := getFluxVersion(fluxConfigMap.Data)
	if err != nil {
		a.logger.Error(err, "I was not able to determine a flux release with respect to the version you defined. Check the configmap in the garden cluster for the version.")
		return err
	}

	// --------------------- Flux Installation ----------------------------

	if !existsManagedResource(ctx, a.client, extensionNamespace, constants.ManagedResourceNameFluxInstall) {
		// Create the resource for the flux installation
		shootResourceFluxInstall, err := createShootResourceFluxInstall(fluxVersion)
		if err != nil {
			return err
		}

		// deploy the managed resource for the flux installatation
		err = managedresources.CreateForShoot(ctx, a.client, extensionNamespace, constants.ManagedResourceNameFluxInstall, true, shootResourceFluxInstall)
		if err != nil {
			return err
		}

		tenMinutes := 10 * time.Minute
		timeoutShootCtx, cancelShootCtx := context.WithTimeout(ctx, tenMinutes)
		defer cancelShootCtx()
		err = managedresources.WaitUntilHealthy(timeoutShootCtx, a.client, extensionNamespace, constants.ManagedResourceNameFluxInstall)
		if err != nil {
			a.logger.Error(err, "I was not able to get the manages resource for the flux installation healthy within 10 minutes. I will delete the manged resource now. You can have another try by reconciling your shoot.")

			managedresources.SetKeepObjects(ctx, a.client, extensionNamespace, constants.ManagedResourceNameFluxInstall, false)
			managedresources.DeleteForShoot(ctx, a.client, extensionNamespace, constants.ManagedResourceNameFluxInstall)

			return err
		}

		// Annotate the managed resource for the flux installation with resources.gardener.cloud/ignore=true
		// This enables the user of the shoot to modify the flux resources, which should be allowed in general,
		// as the extension would be too restrictive otherwise
		retry.Until(ctx, time.Second, func(ctx context.Context) (done bool, err error) {

			if err := setAnnotationMrFluxInstall(ctx, a.client, extensionNamespace); err != nil {
				return retry.MinorError(err)
			}
			return retry.Ok()
		})
	}

	// --------------------- Flux Configuration----------------------------

	if !existsManagedResource(ctx, a.client, extensionNamespace, constants.ManagedResourceNameFluxConfig) {
		// create the resources for the flux configuration
		shootResourceFluxConfig, err := a.createShootResourceFluxConfig(ctx, projectNamespace, fluxConfigMap.Data)
		if err != nil {
			return err
		}
		a.logger.Info("Please add the (public) deploy key to your git repository, you can find it in the secret")

		// deploy the managed resource for the flux configuration
		err = managedresources.CreateForShoot(ctx, a.client, extensionNamespace, constants.ManagedResourceNameFluxConfig, true, shootResourceFluxConfig)
		if err != nil {
			return err
		}
	}
	// return whether an error ocurred
	return err
}

// Delete the Extension resource.
//
// PARAMETERS
// ctx context.Context               Execution context
// ex  *extensionsv1alpha1.Extension Extension struct
func (a *actuator) Delete(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()
	twoMinutes := 2 * time.Minute

	timeoutShootCtx, cancelShootCtx := context.WithTimeout(ctx, twoMinutes)
	defer cancelShootCtx()

	// also delete the objects in case the extension resource is deleted
	if err := managedresources.SetKeepObjects(ctx, a.client, ex.GetNamespace(), constants.ManagedResourceNameFluxInstall, false); err != nil {
		return err
	}
	if err := managedresources.SetKeepObjects(ctx, a.client, ex.GetNamespace(), constants.ManagedResourceNameFluxConfig, false); err != nil {
		return err
	}

	// delete the flux configuration resource
	if err := managedresources.DeleteForShoot(ctx, a.client, namespace, constants.ManagedResourceNameFluxConfig); err != nil {
		return err
	}

	if err := managedresources.WaitUntilDeleted(timeoutShootCtx, a.client, namespace, constants.ManagedResourceNameFluxConfig); err != nil {
		return err
	}

	// now delete the flux installation
	if err := managedresources.DeleteForShoot(ctx, a.client, namespace, constants.ManagedResourceNameFluxInstall); err != nil {
		return err
	}

	if err := managedresources.WaitUntilDeleted(timeoutShootCtx, a.client, namespace, constants.ManagedResourceNameFluxInstall); err != nil {
		return err
	}

	return nil
}

// Restore the Extension resource.
//
// PARAMETERS
// ctx context.Context               Execution context
// ex  *extensionsv1alpha1.Extension Extension struct
func (a *actuator) Restore(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, ex)
}

// Migrate the Extension resource.
//
// PARAMETERS
// ctx context.Context               Execution context
// ex  *extensionsv1alpha1.Extension Extension struct
func (a *actuator) Migrate(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	return a.Delete(ctx, ex)
}

// InjectConfig injects the rest config to this actuator.
//
// PARAMETERS
// config *rest.Config Config to be injected
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config
	return nil
}

// InjectClient injects the controller runtime client into the reconciler.
//
// PARAMETERS
// client client.Client Client to be injected
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	clientInterface, err := gardenclient.NewClientFromSecret(context.Background(), a.client, "garden", "gardenlet-kubeconfig")
	if err != nil {
		return err
	}
	clientInterface.Start(context.Background())
	a.clientGardenlet = clientInterface.Client()
	return nil
}

// InjectScheme injects the given scheme into the reconciler.
//
// PARAMETERS
// scheme *runtime.Scheme Scheme to be injected
func (a *actuator) InjectScheme(scheme *runtime.Scheme) error {
	a.decoder = serializer.NewCodecFactory(scheme, serializer.EnableStrict).UniversalDecoder()
	return nil
}

func createShootResourceFluxInstall(fluxVersion string) (map[string][]byte, error) {

	fluxInstallYaml, err := getFluxInstallYaml(fluxVersion)
	if err != nil {
		return nil, err
	}

	shootResources := make(map[string][]byte)
	shootResources["flux-install-yaml"] = fluxInstallYaml

	return shootResources, nil
}

// createShootResourceFluxConfig ...
func (a *actuator) createShootResourceFluxConfig(ctx context.Context, projectNamespace string, fluxconfig map[string]string) (map[string][]byte, error) {

	fluxSource := getFluxSourceData(fluxconfig)
	fluxKustomization := getFluxKustomizationData()

	shootResources := make(map[string][]byte)

	if fluxconfig["repositoryType"] == "private" {

		var fluxRepoSecretData []byte
		var err error
		fluxRepoSecret := corev1.Secret{}

		// First, we need to check whether the source secret already exists in the projectNamespace.
		// If so, copy the data over to the per shoot secret data. Otherwise, create a new secret and
		// deploy it to the projectNamespace and use it for the managed resource.
		if a.clientGardenlet.Get(ctx, kutil.Key(projectNamespace, constants.FluxSourceSecretName), &fluxRepoSecret) == nil {
			fluxRepoSecret.APIVersion = "v1"
			fluxRepoSecret.Kind = "Secret"
			fluxRepoSecret.ObjectMeta = metav1.ObjectMeta{
				Name:      constants.FluxSourceSecretName,
				Namespace: "flux-system",
			}
			fluxRepoSecretData, err = yaml.Marshal(fluxRepoSecret)
			if err != nil {
				return nil, err
			}
		} else {
			// parse the repository url in order to extract the hostname
			// which is required for the generation of an ssh keypair
			repourl, err := url.Parse(fluxconfig["repositoryUrl"])
			if err != nil {
				return nil, err
			}
			// define some options for the generation of the flux source secret
			sourceSecOpts := sourcesecret.MakeDefaultOptions()
			sourceSecOpts.PrivateKeyAlgorithm = "ed25519"
			sourceSecOpts.SSHHostname = repourl.Hostname()
			sourceSecOpts.Name = constants.FluxSourceSecretName

			// generate the flux source secret manifest and store it as []byte in the shootResources
			secManifest, err := sourcesecret.Generate(sourceSecOpts)
			fluxRepoSecretData = []byte(secManifest.Content)

			// lastly, also deploy the flux source secret into the projectNamespace in the seed cluster
			// in order to reuse it, when other shoots are created
			yaml.Unmarshal(fluxRepoSecretData, &fluxRepoSecret)
			fluxRepoSecret.SetNamespace(projectNamespace)
			a.clientGardenlet.Create(ctx, &fluxRepoSecret)
		}

		shootResources["flux-reposecret"] = fluxRepoSecretData

		// now let's introduce the secret reference into the flux source resource
		fluxSource.Spec.SecretRef = &meta.LocalObjectReference{Name: constants.FluxSourceSecretName}
	}

	fluxSourceData, err := yaml.Marshal(fluxSource)
	if err != nil {
		return nil, err
	}

	fluxKustomizationData, err := yaml.Marshal(fluxKustomization)
	if err != nil {
		return nil, err
	}

	shootResources["flux-source"] = fluxSourceData
	shootResources["flux-ks"] = fluxKustomizationData

	return shootResources, nil
}

// getFluxSourceSecrets ...
func getFluxSourceData(fluxconfig map[string]string) sourcecontrollerv1beta2.GitRepository {

	gitrepo := sourcecontrollerv1beta2.GitRepository{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "source.toolkit.fluxcd.io/v1beta2",
			Kind:       "GitRepository",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FluxGitRepositoryName,
			Namespace: "flux-system",
		},
		Spec: sourcecontrollerv1beta2.GitRepositorySpec{
			Interval: metav1.Duration{
				Duration: time.Second * 30,
			},
			Reference: &sourcecontrollerv1beta2.GitRepositoryRef{
				Branch: fluxconfig["repositoryBranch"],
			},
			URL: fluxconfig["repositoryUrl"],
		},
		Status: sourcecontrollerv1beta2.GitRepositoryStatus{},
	}

	return gitrepo
}

// getFluxSourceSecrets ...
func getFluxKustomizationData() kustomizecontrollerv1beta2.Kustomization {

	ks := kustomizecontrollerv1beta2.Kustomization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kustomize.toolkit.fluxcd.io/v1beta2",
			Kind:       "Kustomization",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FluxMainKustomizationName,
			Namespace: "flux-system",
		},
		Spec: kustomizecontrollerv1beta2.KustomizationSpec{
			Interval: metav1.Duration{
				Duration: time.Minute * 5,
			},
			Path:  "./",
			Prune: true,
			SourceRef: kustomizecontrollerv1beta2.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: constants.FluxGitRepositoryName,
			},
			TargetNamespace: "default",
		},
		Status: kustomizecontrollerv1beta2.KustomizationStatus{},
	}

	return ks
}

// getFluxVersion ...
func getFluxVersion(fluxconfig map[string]string) (string, error) {

	// If the ConfigMap in the garden cluster defines a version, use this version.
	// Otherwise, simply use the latest version available on Github
	var fluxVersion string
	var exists bool
	ghClient := github.NewClient(nil)
	if fluxVersion, exists = fluxconfig["fluxVersion"]; !exists {
		ghReleaseLatest, _, err := ghClient.Repositories.GetLatestRelease(context.Background(), "fluxcd", "flux2")
		if err != nil {
			return "", err
		}
		fluxVersion = *ghReleaseLatest.Name
	} else {
		// Check if the release defined in the ConfigMap exists. If not, return an error
		_, _, err := ghClient.Repositories.GetReleaseByTag(context.Background(), "fluxcd", "flux2", fluxVersion)
		if err != nil {
			return "", err
		}
	}
	return fluxVersion, nil
}

func getFluxInstallYaml(fluxVersion string) ([]byte, error) {

	// download flux install.yaml
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get("https://github.com/fluxcd/flux2/releases/download/" + fluxVersion + "/install.yaml")
	if err != nil {
		return nil, err
	}

	fluxyaml, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return fluxyaml, nil
}

// setAnnotationMrFluxInstall ...
func setAnnotationMrFluxInstall(ctx context.Context, c client.Client, extensionNamespace string) error {

	mrFluxInstall := resourcesv1alpha1.ManagedResource{}
	err := c.Get(ctx, kutil.Key(extensionNamespace, constants.ManagedResourceNameFluxInstall), &mrFluxInstall)
	if err != nil {
		return err
	}

	// Set the annotation
	mrFluxInstall.Annotations = map[string]string{"resources.gardener.cloud/ignore": "true"}

	err = c.Update(ctx, &mrFluxInstall)
	if err != nil {
		return err
	}
	return err
}

// existsManagedResource ...
func existsManagedResource(ctx context.Context, c client.Client, extensionNamespace string, name string) bool {
	mr := resourcesv1alpha1.ManagedResource{}
	err := c.Get(ctx, kutil.Key(extensionNamespace, name), &mr)
	if err != nil {
		return false
	}
	return true
}
