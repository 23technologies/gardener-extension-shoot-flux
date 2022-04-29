# [Gardener Extension for Flux](https://gardener.cloud)

[![reuse compliant](https://reuse.software/badge/reuse-compliant.svg)](https://reuse.software/)

[Project Gardener](https://github.com/gardener/gardener) implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service.
Its main principle is to leverage Kubernetes concepts for all of its tasks.

This controller implements Gardener's extension contract for the `shoot-flux` extension.
An example for a `ControllerRegistration` resource that can be used to register this controller to Gardener can be found [here](example/controller-registration.yaml).

Please find more information regarding the extensibility concepts and a detailed proposal [here](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md).

# What does this package provide?

The general idea of this controller is to install the [fluxcd](https://fluxcd.io/) controllers together with a [flux gitrepository resource](https://fluxcd.io/docs/components/source/gitrepositories/) and a [flux kustomization resource](https://fluxcd.io/docs/components/kustomize/kustomization/) into newly created shoot clusters.
In consequence, your fresh shoot cluster will be reconciled to the state defined in the Git repository by the fluxcd controllers.
Thus, this extension provides a general approach to install addons to shoot clusters.

# How to...

## Use it as a gardener operator
Of course, you need to apply the `controller-registration` resources to the garden cluster first.
Moreover, you will need some configuration pointing to the git repository you want to use as a basis for flux.
This configuration is provided on a per-project basis via a `ConfigMap`:
``` yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: flux-config
  namespace: YOUR-PROJECT-NAMESPACE
data:
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
  * Place the kubeconfig of the seed cluster in `PATH-TO-REPO-ROOT/dev/kubeconfig`
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
