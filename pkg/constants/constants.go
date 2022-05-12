// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package constants defines constants used
package constants

const (
	// ExtensionType is the name of the extension type.
	ExtensionType = "shoot-flux"
	// ServiceName is the name of the service.
	ServiceName = ExtensionType

	extensionServiceName = "extension-" + ServiceName

	// ManagedResourceNamesShoot is the name used to describe the managed shoot resources.
	ManagedResourceNameFluxInstall = extensionServiceName + "-flux-install"
	ManagedResourceNameFluxConfig = extensionServiceName + "-flux-config"
	FluxGitRepositoryName = "main-gitrepo"
	FluxMainKustomizationName = "main-ks"

	// FluxSourceSecretName name of the secret containing the flux sources
	FluxSourceSecretName = "flux-source"
)
