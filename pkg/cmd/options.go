// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package cmd provides Kubernetes controller configuration structures used for command execution
package cmd

import (
	"os"

	extensionscmdcontroller "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionshealthcheckcontroller "github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	extensionsheartbeatcmd "github.com/gardener/gardener/extensions/pkg/controller/heartbeat/cmd"

	"github.com/stackitcloud/gardener-extension-shoot-flux/pkg/controller/healthcheck"
	"github.com/stackitcloud/gardener-extension-shoot-flux/pkg/controller/lifecycle"
)

// ExtensionName is the name of the extension.
const ExtensionName = "shoot-flux"

// Options holds configuration passed to the Shoot Flux controller.
type Options struct {
	generalOptions     *extensionscmdcontroller.GeneralOptions
	restOptions        *extensionscmdcontroller.RESTOptions
	managerOptions     *extensionscmdcontroller.ManagerOptions
	controllerOptions  *extensionscmdcontroller.ControllerOptions
	lifecycleOptions   *extensionscmdcontroller.ControllerOptions
	healthOptions      *extensionscmdcontroller.ControllerOptions
	controllerSwitches *extensionscmdcontroller.SwitchOptions
	reconcileOptions   *extensionscmdcontroller.ReconcilerOptions
	heartbeatOptions   *extensionsheartbeatcmd.Options
	optionAggregator   extensionscmdcontroller.OptionAggregator
}

// NewOptions creates a new Options instance.
func NewOptions() *Options {
	options := &Options{
		generalOptions: &extensionscmdcontroller.GeneralOptions{},
		restOptions:    &extensionscmdcontroller.RESTOptions{},
		managerOptions: &extensionscmdcontroller.ManagerOptions{
			// These are default values.
			LeaderElection:          true,
			LeaderElectionID:        extensionscmdcontroller.LeaderElectionNameID(ExtensionName),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
		},
		controllerOptions: &extensionscmdcontroller.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		lifecycleOptions: &extensionscmdcontroller.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		healthOptions: &extensionscmdcontroller.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		heartbeatOptions: &extensionsheartbeatcmd.Options{
			ExtensionName: ExtensionName,
			Namespace:     os.Getenv("LEADER_ELECTION_NAMESPACE"),
		},
		reconcileOptions: &extensionscmdcontroller.ReconcilerOptions{},
		controllerSwitches: extensionscmdcontroller.NewSwitchOptions(
			extensionscmdcontroller.Switch(lifecycle.Name, lifecycle.AddToManager),
			extensionscmdcontroller.Switch(extensionshealthcheckcontroller.ControllerName, healthcheck.AddToManager),
			extensionscmdcontroller.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager)),
	}

	options.optionAggregator = extensionscmdcontroller.NewOptionAggregator(
		options.generalOptions,
		options.restOptions,
		options.managerOptions,
		options.controllerOptions,
		extensionscmdcontroller.PrefixOption("lifecycle-", options.lifecycleOptions),
		extensionscmdcontroller.PrefixOption("healthcheck-", options.healthOptions),
		extensionscmdcontroller.PrefixOption("heartbeat-", options.heartbeatOptions),
		options.controllerSwitches,
		options.reconcileOptions,
	)

	return options
}
