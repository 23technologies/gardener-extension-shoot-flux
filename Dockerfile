# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

############# builder
FROM golang:1.17.8 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-shoot-flux
COPY . .
RUN make install

############# gardener-extension-shoot-flux
FROM alpine:3.15.0 AS gardener-extension-shoot-flux

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-shoot-flux /gardener-extension-shoot-flux
ENTRYPOINT ["/gardener-extension-shoot-flux"]
