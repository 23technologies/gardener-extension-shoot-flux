// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"

	"github.com/gardener/gardener/cmd/utils"
	"github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	gardenerhealthz "github.com/gardener/gardener/pkg/healthz"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/component-base/version/verflag"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/stackitcloud/gardener-extension-shoot-flux/pkg/controller/extension"
	"github.com/stackitcloud/gardener-extension-shoot-flux/pkg/controller/healthcheck"
)

// Name is a const for the name of this component.
const Name = "gardener-extension-shoot-flux"

// NewCommand creates a new cobra.Command for running gardener-extension-shoot-flux.
func NewCommand() *cobra.Command {
	opts := newOptions()

	cmd := &cobra.Command{
		Use:   Name,
		Short: Name + " bootstraps Flux in shoot clusters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			log, err := utils.InitRun(cmd, opts, Name)
			if err != nil {
				return err
			}
			return run(cmd.Context(), log, opts)
		},
	}

	flags := cmd.Flags()
	verflag.AddFlags(flags)
	opts.addFlags(flags)

	return cmd
}

func run(ctx context.Context, log logr.Logger, o *options) error {
	log.Info("Setting up manager")
	mgr, err := manager.New(o.RESTConfig, o.ManagerOptions)
	if err != nil {
		return err
	}

	log.Info("Setting up health check endpoints")
	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("failed adding ping healthzcheck: %w", err)
	}
	if err := mgr.AddReadyzCheck("informer-sync", gardenerhealthz.NewCacheSyncHealthz(mgr.GetCache())); err != nil {
		return fmt.Errorf("failed adding informer-sync readycheck: %w", err)
	}

	log.Info("Adding controllers to manager")
	o.extensionOptions.Completed().Apply(&extension.DefaultAddOptions.Controller)
	o.healthOptions.Completed().Apply(&healthcheck.DefaultAddOptions.Controller)
	o.heartbeatOptions.Completed().Apply(&heartbeat.DefaultAddOptions)

	if err := o.controllerSwitches.Completed().AddToManager(ctx, mgr); err != nil {
		return fmt.Errorf("failed adding controllers to manager: %w", err)
	}

	log.Info("Starting manager")
	return mgr.Start(ctx)
}
