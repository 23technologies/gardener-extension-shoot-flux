package v1alpha1

const (
	// ConditionBootstrapped is an annotation on the Flux installation namespace that is set by the extension after
	// successfully bootstrapping Flux once. It is used for skipping reconciliation of the Flux resources after a first
	// initial bootstrapping.
	ConditionBootstrapped = "FluxBootstrapped"
)
