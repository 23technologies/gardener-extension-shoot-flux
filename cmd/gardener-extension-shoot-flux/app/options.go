// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"os"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/gardener/gardener/cmd/utils"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionscmdcontroller "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionshealthcheckcontroller "github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	extensionsheartbeatcmd "github.com/gardener/gardener/extensions/pkg/controller/heartbeat/cmd"
	"github.com/gardener/gardener/extensions/pkg/util"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	componentbaseconfig "k8s.io/component-base/config"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	fluxv1alpha1 "github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis/flux/v1alpha1"
	"github.com/stackitcloud/gardener-extension-shoot-flux/pkg/controller/extension"
	"github.com/stackitcloud/gardener-extension-shoot-flux/pkg/controller/healthcheck"
)

var _ utils.Options = &options{}

// options holds configuration passed to the Shoot Flux controller.
type options struct {
	generalOptions     *extensionscmdcontroller.GeneralOptions
	restOptions        *extensionscmdcontroller.RESTOptions
	managerOptions     *extensionscmdcontroller.ManagerOptions
	extensionOptions   *extensionscmdcontroller.ControllerOptions
	healthOptions      *extensionscmdcontroller.ControllerOptions
	heartbeatOptions   *extensionsheartbeatcmd.Options
	controllerSwitches *extensionscmdcontroller.SwitchOptions
	reconcileOptions   *extensionscmdcontroller.ReconcilerOptions
	optionAggregator   extensionscmdcontroller.OptionAggregator

	// completed options
	RESTConfig     *rest.Config
	ManagerOptions manager.Options
}

// newOptions creates a new options instance.
func newOptions() *options {
	opts := &options{
		generalOptions: &extensionscmdcontroller.GeneralOptions{},
		restOptions:    &extensionscmdcontroller.RESTOptions{},
		managerOptions: &extensionscmdcontroller.ManagerOptions{
			// These are default values.
			LeaderElection:          true,
			LeaderElectionID:        extensionscmdcontroller.LeaderElectionNameID(fluxv1alpha1.ExtensionType),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
			MetricsBindAddress:      ":8080",
			HealthBindAddress:       ":8081",
		},
		extensionOptions: &extensionscmdcontroller.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		healthOptions: &extensionscmdcontroller.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		heartbeatOptions: &extensionsheartbeatcmd.Options{
			ExtensionName: fluxv1alpha1.ExtensionType,
			Namespace:     os.Getenv("LEADER_ELECTION_NAMESPACE"),
		},
		controllerSwitches: extensionscmdcontroller.NewSwitchOptions(
			extensionscmdcontroller.Switch(extension.ControllerName, extension.AddToManager),
			extensionscmdcontroller.Switch(extensionshealthcheckcontroller.ControllerName, healthcheck.AddToManager),
			extensionscmdcontroller.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
		),
		reconcileOptions: &extensionscmdcontroller.ReconcilerOptions{},
	}

	opts.optionAggregator = extensionscmdcontroller.NewOptionAggregator(
		opts.generalOptions,
		opts.restOptions,
		opts.managerOptions,
		extensionscmdcontroller.PrefixOption(extension.ControllerName+"-", opts.extensionOptions),
		extensionscmdcontroller.PrefixOption(extensionshealthcheckcontroller.ControllerName+"-", opts.healthOptions),
		extensionscmdcontroller.PrefixOption(extensionsheartbeatcontroller.ControllerName+"-", opts.heartbeatOptions),
		opts.controllerSwitches,
		opts.reconcileOptions,
	)

	return opts
}

func (o *options) addFlags(fs *pflag.FlagSet) {
	o.optionAggregator.AddFlags(fs)
}

func (o *options) Complete() error {
	if err := o.optionAggregator.Complete(); err != nil {
		return err
	}

	// customize rest config
	o.RESTConfig = o.restOptions.Completed().Config
	// TODO: consider dropping this or make these settings configurable
	util.ApplyClientConnectionConfigurationToRESTConfig(&componentbaseconfig.ClientConnectionConfiguration{
		QPS:   100.0,
		Burst: 130,
	}, o.RESTConfig)

	// customize manager options
	o.ManagerOptions = o.managerOptions.Completed().Options()
	o.ManagerOptions.WebhookServer = nil // this extension doesn't run webhooks
	o.ManagerOptions.GracefulShutdownTimeout = ptr.To(5 * time.Second)

	o.ManagerOptions.Client.Cache = &client.CacheOptions{
		DisableFor: []client.Object{
			&corev1.Secret{}, // applied for ManagedResources
		},
	}

	scheme := runtime.NewScheme()
	if err := (&runtime.SchemeBuilder{
		extensionscontroller.AddToScheme,
		fluxv1alpha1.AddToScheme,
		apiextensionsv1.AddToScheme,
		sourcev1.AddToScheme,
		kustomizev1.AddToScheme,
	}).AddToScheme(scheme); err != nil {
		return err
	}
	o.ManagerOptions.Scheme = scheme

	return nil
}

func (o *options) Validate() error {
	if err := o.heartbeatOptions.Validate(); err != nil {
		return err
	}

	return nil
}

func (o *options) LogConfig() (logLevel, logFormat string) {
	return o.managerOptions.LogLevel, o.managerOptions.LogFormat
}
