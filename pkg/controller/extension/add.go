package extension

import (
	"context"
	"time"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	fluxv1alpha1 "github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis/flux/v1alpha1"
)

const (
	// ControllerName is the name of the controller.
	ControllerName = "extension"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the extension controller to the manager.
type AddOptions struct {
	// Controller are the controller.Options.
	Controller controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
func AddToManagerWithOptions(ctx context.Context, mgr manager.Manager, opts AddOptions) error {
	return extension.Add(ctx, mgr, extension.AddArgs{
		Actuator:          NewActuator(mgr.GetClient()),
		ControllerOptions: opts.Controller,
		Name:              ControllerName,
		FinalizerSuffix:   fluxv1alpha1.ExtensionType,
		Resync:            60 * time.Minute,
		Predicates:        extension.DefaultPredicates(ctx, mgr, opts.IgnoreOperationAnnotation),
		Type:              fluxv1alpha1.ExtensionType,
	})
}

// AddToManager adds a controller with the default Options.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return AddToManagerWithOptions(ctx, mgr, DefaultAddOptions)
}
