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

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

// Actuator acts upon Extension resources.
type Actuator interface {
	// Reconcile the Extension resource.
	Reconcile(ctx context.Context, ex *extensionsv1alpha1.Extension) error
	// Delete the Extension resource.
	Delete(ctx context.Context, ex *extensionsv1alpha1.Extension) error
	// Restore the Extension resource.
	Restore(ctx context.Context, ex *extensionsv1alpha1.Extension) error
	// Migrate the Extension resource.
	Migrate(ctx context.Context, ex *extensionsv1alpha1.Extension) error
}
