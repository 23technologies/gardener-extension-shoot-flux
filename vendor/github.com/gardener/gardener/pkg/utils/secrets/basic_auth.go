// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package secrets

import (
	"github.com/gardener/gardener/pkg/utils"
)

type formatType string

const (
	// BasicAuthFormatNormal indicates that the data map should be rendered the normal way (dedicated keys for
	// username and password.
	BasicAuthFormatNormal formatType = "normal"

	// DataKeyUserName is the key in a secret data holding the username.
	DataKeyUserName = "username"
	// DataKeyPassword is the key in a secret data holding the password.
	DataKeyPassword = "password"
	// DataKeySHA1Auth is the key in a secret data holding the sha1-schemed credentials pair as string.
	DataKeySHA1Auth = "auth"
)

// BasicAuthSecretConfig contains the specification for a to-be-generated basic authentication secret.
type BasicAuthSecretConfig struct {
	Name string
	// Format is the format type.
	//
	// Do not remove this field, even though the field is not used and there is only one supported format ("normal").
	// The secret manager computes the Secret hash based on the config object (BasicAuthSecretConfig). A field removal in the
	// BasicAuthSecretConfig object would compute a new Secret hash and this would lead the existing Secrets to be regenerated.
	// Hence, usages of the BasicAuthSecretConfig should continue to pass the Format field with value "normal".
	Format formatType

	Username       string
	PasswordLength int
}

// BasicAuth contains the username, the password and optionally hash of the password.
type BasicAuth struct {
	Name string

	Username string
	Password string
}

// GetName returns the name of the secret.
func (s *BasicAuthSecretConfig) GetName() string {
	return s.Name
}

// Generate implements ConfigInterface.
func (s *BasicAuthSecretConfig) Generate() (DataInterface, error) {
	password, err := GenerateRandomString(s.PasswordLength)
	if err != nil {
		return nil, err
	}

	basicAuth := &BasicAuth{
		Name: s.Name,

		Username: s.Username,
		Password: password,
	}

	return basicAuth, nil
}

// SecretData computes the data map which can be used in a Kubernetes secret.
func (b *BasicAuth) SecretData() map[string][]byte {
	data := map[string][]byte{}

	data[DataKeyUserName] = []byte(b.Username)
	data[DataKeyPassword] = []byte(b.Password)
	data[DataKeySHA1Auth] = utils.CreateSHA1Secret(data[DataKeyUserName], data[DataKeyPassword])

	return data
}
