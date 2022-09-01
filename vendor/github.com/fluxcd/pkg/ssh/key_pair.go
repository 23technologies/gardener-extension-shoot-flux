/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ssh

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// KeyPair holds the public and private key PEM block bytes.
type KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

type KeyPairGenerator interface {
	Generate() (*KeyPair, error)
}

type KeyPairType string

const (
	// RSA_4096 represents a RSA keypair with 4096 bits.
	RSA_4096 KeyPairType = "rsa-4096"
	// ECDSA_P256 represents a ecdsa keypair using Curve P-256.
	ECDSA_P256 KeyPairType = "ecdsa-p256"
	// ECDSA_P384 represents a ecdsa keypair using Curve P-384.
	ECDSA_P384 KeyPairType = "ecdsa-p384"
	// ECDSA_P521 represents a ecdsa keypair using Curve P-521.
	ECDSA_P521 KeyPairType = "ecdsa-p521"
	// ED25519 represents a ed25519 keypair.
	ED25519 KeyPairType = "ed25519"
)

// GenerateKeyPair generates a keypair based on KeyPairType.
func GenerateKeyPair(keyType KeyPairType) (*KeyPair, error) {
	switch keyType {
	case RSA_4096:
		return NewRSAGenerator(4096).Generate()
	case ECDSA_P256:
		return NewECDSAGenerator(elliptic.P256()).Generate()
	case ECDSA_P384:
		return NewECDSAGenerator(elliptic.P384()).Generate()
	case ECDSA_P521:
		return NewECDSAGenerator(elliptic.P521()).Generate()
	case ED25519:
		return NewEd25519Generator().Generate()
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

type RSAGenerator struct {
	bits int
}

func NewRSAGenerator(bits int) KeyPairGenerator {
	return &RSAGenerator{bits}
}

func (g *RSAGenerator) Generate() (*KeyPair, error) {
	pk, err := rsa.GenerateKey(rand.Reader, g.bits)
	if err != nil {
		return nil, err
	}
	err = pk.Validate()
	if err != nil {
		return nil, err
	}
	pub, err := generatePublicKey(&pk.PublicKey)
	if err != nil {
		return nil, err
	}
	priv, err := encodePrivateKeyToPEM(pk)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

type ECDSAGenerator struct {
	c elliptic.Curve
}

func NewECDSAGenerator(c elliptic.Curve) KeyPairGenerator {
	return &ECDSAGenerator{c}
}

func (g *ECDSAGenerator) Generate() (*KeyPair, error) {
	pk, err := ecdsa.GenerateKey(g.c, rand.Reader)
	if err != nil {
		return nil, err
	}
	pub, err := generatePublicKey(&pk.PublicKey)
	if err != nil {
		return nil, err
	}
	priv, err := encodePrivateKeyToPEM(pk)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

type Ed25519Generator struct{}

func NewEd25519Generator() KeyPairGenerator {
	return &Ed25519Generator{}
}

func (g *Ed25519Generator) Generate() (*KeyPair, error) {
	pk, pv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	pub, err := generatePublicKey(pk)
	if err != nil {
		return nil, err
	}
	priv, err := encodePrivateKeyToPEM(pv)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

func generatePublicKey(pk interface{}) ([]byte, error) {
	b, err := ssh.NewPublicKey(pk)
	if err != nil {
		return nil, err
	}
	k := ssh.MarshalAuthorizedKey(b)
	return k, nil
}

// encodePrivateKeyToPEM encodes the given private key to a PEM block.
// The encoded format is PKCS#8 for universal support of the most
// common key types (rsa, ecdsa, ed25519).
func encodePrivateKeyToPEM(pk interface{}) ([]byte, error) {
	b, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return nil, err
	}
	block := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}
	return pem.EncodeToMemory(&block), nil
}
