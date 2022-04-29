// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package extension

import (
	"context"
	"fmt"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/controllerutils"
	reconcilerutils "github.com/gardener/gardener/pkg/controllerutils/reconciler"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// reconciler reconciles Extension resources of Gardener's
// `extensions.gardener.cloud` API group.
type reconciler struct {
	logger   logr.Logger
	actuator Actuator

	client        client.Client
	reader        client.Reader
	statusUpdater extensionscontroller.StatusUpdater

	resync        time.Duration
	finalizerName string
}

// NewReconciler creates a new reconcile.Reconciler that reconciles
// Extension resources of Gardener's `extensions.gardener.cloud` API group.
func NewReconciler(args AddArgs) reconcile.Reconciler {
	logger := log.Log.WithName(args.Name)

	return reconcilerutils.OperationAnnotationWrapper(
		func() client.Object { return &extensionsv1alpha1.Extension{} },
		&reconciler{
			logger:        logger,
			actuator:      args.Actuator,
			statusUpdater: extensionscontroller.NewStatusUpdater(logger),
			finalizerName: fmt.Sprintf("%s/%s", FinalizerPrefix, args.FinalizerSuffix),
			resync:        args.Resync,
		},
	)
}

// InjectFunc enables dependency injection into the actuator.
func (r *reconciler) InjectFunc(f inject.Func) error {
	return f(r.actuator)
}

// InjectClient injects the controller runtime client into the reconciler.
func (r *reconciler) InjectClient(client client.Client) error {
	r.client = client
	r.statusUpdater.InjectClient(client)
	return nil
}

func (r *reconciler) InjectAPIReader(reader client.Reader) error {
	r.reader = reader
	return nil
}

// Reconcile is the reconciler function that gets executed in case there are new events for `Extension` resources.
func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	ex := &extensionsv1alpha1.Extension{}
	if err := r.client.Get(ctx, request.NamespacedName, ex); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	var result reconcile.Result

	cluster, err := extensionscontroller.GetCluster(ctx, r.client, ex.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	logger := r.logger.WithValues("extension", kutil.ObjectName(ex))
	if extensionscontroller.IsFailed(cluster) {
		logger.Info("Skipping the reconciliation of extension of failed shoot")
		return reconcile.Result{}, nil
	}

	operationType := gardencorev1beta1helper.ComputeOperationType(ex.ObjectMeta, ex.Status.LastOperation)

	if cluster.Shoot != nil && operationType != gardencorev1beta1.LastOperationTypeMigrate {
		key := "extension:" + kutil.ObjectName(ex)
		ok, watchdogCtx, cleanup, err := common.GetOwnerCheckResultAndContext(ctx, r.client, ex.Namespace, cluster.Shoot.Name, key)
		if err != nil {
			return reconcile.Result{}, err
		} else if !ok {
			return reconcile.Result{}, fmt.Errorf("this seed is not the owner of shoot %s", kutil.ObjectName(cluster.Shoot))
		}
		ctx = watchdogCtx
		if cleanup != nil {
			defer cleanup()
		}
	}

	switch {
	case extensionscontroller.ShouldSkipOperation(operationType, ex):
		return reconcile.Result{}, nil
	case operationType == gardencorev1beta1.LastOperationTypeMigrate:
		return r.migrate(ctx, ex)
	case ex.DeletionTimestamp != nil:
		return r.delete(ctx, ex)
	case operationType == gardencorev1beta1.LastOperationTypeRestore:
		return r.restore(ctx, ex, operationType)
	default:
		if result, err = r.reconcile(ctx, ex, operationType); err != nil {
			return result, err
		}
		return reconcile.Result{Requeue: r.resync != 0, RequeueAfter: r.resync}, nil
	}
}

func (r *reconciler) reconcile(ctx context.Context, ex *extensionsv1alpha1.Extension, operationType gardencorev1beta1.LastOperationType) (reconcile.Result, error) {
	if err := controllerutils.EnsureFinalizer(ctx, r.reader, r.client, ex, r.finalizerName); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.statusUpdater.Processing(ctx, ex, operationType, "Reconciling the extension"); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Starting the reconciliation of extension", "extension", kutil.ObjectName(ex))
	if err := r.actuator.Reconcile(ctx, ex); err != nil {
		_ = r.statusUpdater.Error(ctx, ex, reconcilerutils.ReconcileErrCauseOrErr(err), operationType, "Error reconciling extension")
		return reconcilerutils.ReconcileErr(err)
	}

	if err := r.statusUpdater.Success(ctx, ex, operationType, "Successfully reconciled extension"); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) delete(ctx context.Context, ex *extensionsv1alpha1.Extension) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(ex, r.finalizerName) {
		r.logger.Info("Deleting extension causes a no-op as there is no finalizer", "extension", kutil.ObjectName(ex))
		return reconcile.Result{}, nil
	}

	if err := r.statusUpdater.Processing(ctx, ex, gardencorev1beta1.LastOperationTypeDelete, "Deleting the extension"); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Starting the deletion of extension", "extension", kutil.ObjectName(ex))
	if err := r.actuator.Delete(ctx, ex); err != nil {
		_ = r.statusUpdater.Error(ctx, ex, reconcilerutils.ReconcileErrCauseOrErr(err), gardencorev1beta1.LastOperationTypeDelete, "Error deleting the extension")
		return reconcilerutils.ReconcileErr(err)
	}

	if err := r.statusUpdater.Success(ctx, ex, gardencorev1beta1.LastOperationTypeDelete, "Successfully deleted the extension"); err != nil {
		return reconcile.Result{}, err
	}
	r.logger.Info("Removing finalizer", "extension", kutil.ObjectName(ex))
	if err := controllerutils.RemoveFinalizer(ctx, r.reader, r.client, ex, r.finalizerName); err != nil {
		return reconcile.Result{}, fmt.Errorf("error removing finalizer from extension: %+v", err)
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) restore(ctx context.Context, ex *extensionsv1alpha1.Extension, operationType gardencorev1beta1.LastOperationType) (reconcile.Result, error) {
	if err := controllerutils.EnsureFinalizer(ctx, r.reader, r.client, ex, r.finalizerName); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.statusUpdater.Processing(ctx, ex, operationType, "Restoring Extension resource"); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Starting the restoration of extension", "extension", kutil.ObjectName(ex))
	if err := r.actuator.Restore(ctx, ex); err != nil {
		_ = r.statusUpdater.Error(ctx, ex, reconcilerutils.ReconcileErrCauseOrErr(err), operationType, "Unable to restore Extension resource")
		return reconcilerutils.ReconcileErr(err)
	}

	if err := r.statusUpdater.Success(ctx, ex, operationType, "Successfully restored Extension resource"); err != nil {
		return reconcile.Result{}, err
	}

	if err := extensionscontroller.RemoveAnnotation(ctx, r.client, ex, v1beta1constants.GardenerOperation); err != nil {
		return reconcile.Result{}, fmt.Errorf("error removing annotation from Extension resource: %+v", err)
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) migrate(ctx context.Context, ex *extensionsv1alpha1.Extension) (reconcile.Result, error) {
	if err := r.statusUpdater.Processing(ctx, ex, gardencorev1beta1.LastOperationTypeMigrate, "Migrate Extension resource."); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Starting the migration of extension", "extension", kutil.ObjectName(ex))
	if err := r.actuator.Migrate(ctx, ex); err != nil {
		_ = r.statusUpdater.Error(ctx, ex, reconcilerutils.ReconcileErrCauseOrErr(err), gardencorev1beta1.LastOperationTypeMigrate, "Error migrating Extension resource")
		return reconcilerutils.ReconcileErr(err)
	}

	if err := r.statusUpdater.Success(ctx, ex, gardencorev1beta1.LastOperationTypeMigrate, "Successfully migrated Extension resource"); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Removing all finalizers", "extension", kutil.ObjectName(ex))
	if err := controllerutils.RemoveAllFinalizers(ctx, r.client, r.client, ex); err != nil {
		return reconcile.Result{}, fmt.Errorf("error removing finalizers from extension: %+v", err)
	}

	if err := extensionscontroller.RemoveAnnotation(ctx, r.client, ex, v1beta1constants.GardenerOperation); err != nil {
		return reconcile.Result{}, fmt.Errorf("error removing annotation from Extension resource: %+v", err)
	}

	return reconcile.Result{}, nil
}
