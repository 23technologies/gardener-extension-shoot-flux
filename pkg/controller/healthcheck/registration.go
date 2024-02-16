// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package healthcheck contains functions used at the healthcheck controller
package healthcheck

import (
	"context"
	"time"

	extensionsconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	defaultSyncPeriod = time.Second * 30
	// DefaultAddOptions contains configuration for the health check controller.
	DefaultAddOptions = healthcheck.DefaultAddArgs{
		HealthCheckConfig: extensionsconfig.HealthCheckConfig{SyncPeriod: metav1.Duration{Duration: defaultSyncPeriod}},
	}
)

// RegisterHealthChecks registers health checks for the Extension resource.
// The controller doesn't actually perform any health checks. However, it removes the health check Conditions written
// by previous versions of the extension from the Extension status.
func RegisterHealthChecks(ctx context.Context, mgr manager.Manager, opts healthcheck.DefaultAddArgs) error {
	// Health checks don't make any sense as long as the extension only cares about bootstrapping flux once.
	// This happens during Extension reconciliation, so the shoot reconciliation either continues successfully or fails at
	// this step.
	// During the cluster's lifetime, the status cannot change as such, as the extension has already fulfilled its
	// purpose.
	// TODO: add real health checks when this extension offers reconciling the flux resources
	return healthcheck.DefaultRegistration(
		ctx,
		"shoot-flux",
		extensionsv1alpha1.SchemeGroupVersion.WithKind(extensionsv1alpha1.ExtensionResource),
		func() client.ObjectList { return &extensionsv1alpha1.ExtensionList{} },
		func() extensionsv1alpha1.Object { return &extensionsv1alpha1.Extension{} },
		mgr,
		opts,
		nil,
		nil,
		sets.New(gardencorev1beta1.ShootSystemComponentsHealthy),
	)
}

// AddToManager adds a controller with the default Options.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return RegisterHealthChecks(ctx, mgr, DefaultAddOptions)
}
