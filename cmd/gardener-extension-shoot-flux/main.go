// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package main provides the application's entry point
package main

import (
	"fmt"

	"github.com/gardener/gardener/pkg/logger"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/stackitcloud/gardener-extension-shoot-flux/pkg/cmd"
)

func main() {
	runtimelog.SetLogger(logger.MustNewZapLogger("", ""))

	ctx := signals.SetupSignalHandler()
	if err := cmd.NewServiceControllerCommand().ExecuteContext(ctx); err != nil {
		fmt.Println(err)
	}
}
