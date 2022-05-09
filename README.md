# Gardener Extension for Flux

[![reuse compliant](https://reuse.software/badge/reuse-compliant.svg)](https://reuse.software/)

[Project Gardener](https://github.com/gardener/gardener) implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service.
Its main principle is to leverage Kubernetes concepts for all of its tasks.

This controller implements Gardener's extension contract for the `shoot-flux` extension.
An example for a `ControllerRegistration` resource that can be used to register this controller to Gardener can be found [here](example/controller-registration.yaml).

Please find more information regarding the extensibility concepts and a detailed proposal [here](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md).

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->
**Table of Contents**

- [Gardener Extension for Flux](#gardener-extension-for-flux)
- [What does this package provide?](#what-does-this-package-provide)
    - [Example use case](#example-use-case)
- [How to...](#how-to)
    - [Use it as a gardener operator](#use-it-as-a-gardener-operator)
    - [Develop this extension locally](#develop-this-extension-locally)
        - [Prerequisites](#prerequisites)
        - [Running and Debugging the controller](#running-and-debugging-the-controller)
- [General concepts and how it works internally](#general-concepts-and-how-it-works-internally)
- [Last remarks](#last-remarks)

<!-- markdown-toc end -->

# What does this package provide?
The general idea of this controller is to install the [fluxcd](https://fluxcd.io/) controllers together with a [flux gitrepository resource](https://fluxcd.io/docs/components/source/gitrepositories/) and a [flux kustomization resource](https://fluxcd.io/docs/components/kustomize/kustomization/) into newly created shoot clusters.
In consequence, your fresh shoot cluster will be reconciled to the state defined in the Git repository by the fluxcd controllers.
Thus, this extension provides a general approach to install addons to shoot clusters.

## Example use case
Let's say you have a CI-workflow which needs a kubernetes cluster with some basic components, such as [cert-manager](https://cert-manager.io/) or [minio](https://min.io/).
Thus, your CI-workflow creates a `Shoot` on which you perform all your actions.
However, the process of creating the `Shoot` and installing the needed components takes for several minutes holding you back from effectively running your CI-pipeline.
In this case, you can make use of this extension and pre-spawn `Shoot`s, which are automatically equipped with fluxcd and reconciled to the state defined in a Git repository.
Of course, there is a trade-off, as your pre-spawned shoots will consume some resources (either in terms of money, if running in a public cloud, or in terms of physical resources).
However, in certain scenarios, this approach will dramatically improve the effectiveness of you CI-workflow.

# How to...

## Use it as a gardener operator
Of course, you need to apply the `controller-registration` resources to the garden cluster first.
You can find the corresponding yaml-files in our [releases](https://github.com/23technologies/gardener-extension-shoot-flux/releases).
Moreover, you will need some configuration pointing to the git repository you want to use as a basis for flux.
This configuration is provided on a per-project basis via a `ConfigMap`:
``` yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: flux-config
  namespace: YOUR-PROJECT-NAMESPACE
data:
  fluxVersion: v0.29.5 # optional, if not defined the latest release will be used
  repositoryUrl: ssh://git@github.com/THE-OWNER/THE-REPO
  repositoryBranch: main
  repositoryType: private
```
At the time writing this, the extension-controller will generate a new SSH-keypair for you, if the `repositoryType` is set to private. This keypair will be stored in the garden cluster as a `Secret` and you will need to make the public key available to you git SSH-server, so that flux can read from your repository. Note: On e.g. github.com, this can be achieved by adding a deploy key to your Git repository.

Next you can deploy a `Shoot` with the `shoot-flux` extension enabled:
``` yaml
apiVersion: core.gardener.cloud/v1beta1
kind: Shoot
metadata:
  name: bar
  namespace: garden-foo
spec:
  extensions:
  - type: shoot-flux
...
```
Then, your shoot cluster should be reconciled to the declarative definition in your Git repository.

## Develop this extension locally
### Prerequisites
  * A local installation of Go
  * A Gardener (could also be the [local setup](https://gardener.cloud/docs/gardener/development/getting_started_locally/))

### Running and Debugging the controller
  * Place the kubeconfig of the `Seed` cluster in `PATH-TO-REPO-ROOT/dev/kubeconfig`
  * Set `ignoreResources=true` and `replicaCount=0` in `PATH-TO-REPO-ROOT/charts/gardener-extension-shoot-flux/values.yaml`
  * Generate the `controller-registration.yaml` by e.g.
  ``` shell
  ./vendor/github.com/gardener/gardener/hack/generate-controller-registration.sh shoot-flux charts/gardener-extension-shoot-flux v0.0.1 example/controller-registration.yaml Extension:shoot-flux
  ```
  * Apply it to the garden cluster:
  ``` shell
  kubectl apply -f PATH-TO-REPO-ROOT/example/controller-registration.yaml
  ```
  * Run and debug the controller with [dlv](https://github.com/go-delve/delve) by:
  ``` shell
  dlv debug ./cmd/gardener-extension-shoot-flux -- --kubeconfig=dev/kubeconfig.yaml  --ignore-operation-annotation=true --leader-election=false --gardener-version="v1.44.4"
  ```
  * You can set breakpoints now, and instruct dlv to run the controller by entering "c" into the dlv commandline.
  * Lastly, deploy a `ConfigMap` pointing to a git repository and a `Shoot` with the `shoot-flux` extension enabled (as explained [above](#use-it-as-a-gardener-operator)).

# General concepts and how it works internally
Generally, this extension was motivated by the idea of enabling Gardener operators to pre-configure `Shoot` clusters with addons.
Obviously, a declarative approach to the configuration makes sense in this scenario.
Consequently fluxcd was chosen as a more general configuration tool for Kubernetes clusters.
From this basis, a Gardener operator can track the configuration of `Shoot` clusters in Git repositories and configure all `Shoot`s in a project to use a configuration via a `ConfigMap` in the project namespace.
This overall workflow is depicted in the block diagram below.

```
                 ┌─────────────────────────────────────────────────────────┐
                 │ Gardener operator                                       │
                 ├─────────────────────────────────────────────────────────┤
                 │ - A human being                                         │
                 │                                                         ├────────────┐
                 │                                                         │            │
                 │                                                         │            │
                 └────────┬────────────────────────────────────────────────┘            │
                          │                           ▲                                 │configures
                          │deploys                    │                                 │SSH-key
                          │Configmap                  │read SSH-key                     │
                          │                           │                                 │
                          ▼                           │                                 │
                 ┌────────────────────────────────────┴───────────────────┐             │
                 │ Garden cluster                                         │             │
                 ├────────────────────────┬─────────────────────────┬─────┤             │
                 │ Projetct 1             │ Project 2               │ ... │             ▼
                 ├────────────────────────┼─────────────────────────┼─────┤  ┌─────────────────────┐
                 │- Configmap containing  │- Configmap containing   │     │  │ Git repository      │
                 │  flux configuration    │  flux configuration     │     │  ├─────────────────────┤
                 │                        │                         │     │  │ - Configuration for │
            ┌───►│- ControllerRegistration│- ControllerRegistration │ ... │  │   shoot clusters    │
            │    │                        │                         │     │  └─────────────────────┘
            │    │- Shoot with extension  │- Shoot with extension   │     │             ▲
            │    │  enabled               │  enabled                │     │             │
            │    │                        │                         │     │             │
read config │    │                        │                         │     │             │
and generate│    └────────────────────────┴─────────────────────────┴─────┘             │reconcile
SSH-keys    │                                                                           │
            │    ┌────────────────────────┐     ┌────────────────────────┐              │
            │    │ Seed cluster           │     │ Shoot cluster          │              │
            │    ├────────────────────────┤     ├────────────────────────┤              │
            │    │- Controller watching   │     │                        │              │
            └────┼─ extension resource    │     │- Flux controllers  ────┼──────────────┘
                 │     │                  │     │                        │
                 │     │deploys           │     │- GitRepository resource│
                 │     │                  │     │                        │
                 │     ▼                  │     │- A main kustomization  │
                 │- Managed resources     │     │                        │
                 │  for flux controllers  │     │                        │
                 │  and flux config       │     │                        │
                 │                        │     │                        │
                 └────────────────────────┘     └────────────────────────┘
```
Wait! How does the controller in the `Seed` cluster communicate to the garden cluster?
Actually, we are just using the `Secret` containing the `gardenlet-kubeconfig` which should be available, when the gardenlet is run inside the `Seed` cluster.
Of course, this is not a rock solid solution, but it was an easy way to achieve the overall goal by simple means.

# Last remarks
This extensions is still in a preliminary state and contains some hacks.
However, the work and testing is still ongoing and the extension will be continuously improved.
In general, you could consider the current state of the extension as kind of a minimal working example for Gardener extensions, as it is very low complex to this point.
