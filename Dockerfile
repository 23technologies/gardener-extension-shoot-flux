# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

############# builder
FROM golang:1.20 AS builder

ENV BINARY_PATH=/go/bin
WORKDIR /go/src/github.com/23technologies/gardener-extension-shoot-flux

COPY . .
RUN make install

############# base
FROM eu.gcr.io/gardener-project/3rd/alpine:3.15 as base

############# gardener-extension-shoot-flux
FROM base AS gardener-extension-shoot-flux
LABEL org.opencontainers.image.source="https://github.com/23technologies/gardener-extension-shoot-flux"

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-shoot-flux /gardener-extension-shoot-flux
ENTRYPOINT ["/gardener-extension-shoot-flux"]
