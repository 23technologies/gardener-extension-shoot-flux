// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/v44/github"
)

var _ = Describe("Flux", func() {

	// check whether the specified version is pulled
	It("Should return the correct version string", func() {
		Expect(getFluxVersion(map[string]string{"fluxVersion": "v0.28.2"})).To(Equal("v0.28.2"))
	})

	// release v0.28.343439 should not exist
	It("Should return an error, as the requested version does not exist", func() {
		_, err := getFluxVersion(map[string]string{"fluxVersion": "v0.28.343439"})
		Expect(err).Should(HaveOccurred())
	})

	// check whether the latest version is pulled in case of zero-conf
	It("Should return the version string of the latest version", func() {
		ghClient := github.NewClient(nil)
		ghReleaseLatest, _, _  := ghClient.Repositories.GetLatestRelease(context.Background(), "fluxcd", "flux2")
		Expect(getFluxVersion(map[string]string{})).To(Equal(*ghReleaseLatest.Name))
	})

	// check whether getFluxInstallYaml returns []byte
	It("Should return the flux-install.yaml", func() {
		Expect(getFluxInstallYaml("v0.28.2")).Should(BeAssignableToTypeOf([]byte{}))
	})

})
