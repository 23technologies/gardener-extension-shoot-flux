<p>Packages:</p>
<ul>
<li>
<a href="#flux.extensions.gardener.cloud%2fv1alpha1">flux.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="flux.extensions.gardener.cloud/v1alpha1">flux.extensions.gardener.cloud/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="flux.extensions.gardener.cloud/v1alpha1.AdditionalResource">AdditionalResource
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.extensions.gardener.cloud/v1alpha1.FluxConfig">FluxConfig</a>)
</p>
<p>
<p>AdditionalResource to sync to the shoot.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name references a resource under Shoot.spec.resources.</p>
</td>
</tr>
<tr>
<td>
<code>targetName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetName optionally overwrites the name of the secret in the shoot.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="flux.extensions.gardener.cloud/v1alpha1.FluxConfig">FluxConfig
</h3>
<p>
<p>FluxConfig specifies how to bootstrap Flux on the shoot cluster.
When both &ldquo;Source&rdquo; and &ldquo;Kustomization&rdquo; are provided they are also installed in the shoot.
Otherwise, only Flux itself is installed with no Objects to reconcile.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>flux</code></br>
<em>
<a href="#flux.extensions.gardener.cloud/v1alpha1.FluxInstallation">
FluxInstallation
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flux configures the Flux installation in the Shoot cluster.</p>
</td>
</tr>
<tr>
<td>
<code>source</code></br>
<em>
<a href="#flux.extensions.gardener.cloud/v1alpha1.Source">
Source
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Source configures how to bootstrap a Flux source object.
If provided, a &ldquo;Kustomization&rdquo; must also be provided.</p>
</td>
</tr>
<tr>
<td>
<code>kustomization</code></br>
<em>
<a href="#flux.extensions.gardener.cloud/v1alpha1.Kustomization">
Kustomization
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Kustomization configures how to bootstrap a Flux Kustomization object.
If provided, &ldquo;Source&rdquo; must also be provided.</p>
</td>
</tr>
<tr>
<td>
<code>additionalSecretResources</code></br>
<em>
<a href="#flux.extensions.gardener.cloud/v1alpha1.AdditionalResource">
[]AdditionalResource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AdditionalSecretResources to sync to the shoot.
Secrets referenced here are only created if they don&rsquo;t exist in the shoot yet.
When a secret is removed from this list, it is deleted in the shoot.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="flux.extensions.gardener.cloud/v1alpha1.FluxInstallation">FluxInstallation
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.extensions.gardener.cloud/v1alpha1.FluxConfig">FluxConfig</a>)
</p>
<p>
<p>FluxInstallation configures the Flux installation in the Shoot cluster.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>version</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Version specifies the Flux version that should be installed.
Defaults to &ldquo;v2.3.0&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>registry</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Registry specifies the container registry where the Flux controller images are pulled from.
Defaults to &ldquo;ghcr.io/fluxcd&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace specifes the namespace where Flux should be installed.
Defaults to &ldquo;flux-system&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="flux.extensions.gardener.cloud/v1alpha1.Kustomization">Kustomization
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.extensions.gardener.cloud/v1alpha1.FluxConfig">FluxConfig</a>)
</p>
<p>
<p>Kustomization configures how to bootstrap a Flux Kustomization object.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>template</code></br>
<em>
<a href="https://fluxcd.io/flux/components/kustomize/api/v1/#kustomize.toolkit.fluxcd.io/v1.Kustomization">
kustomize.toolkit.fluxcd.io/v1.Kustomization
</a>
</em>
</td>
<td>
<p>Template is a partial Kustomization object in API version kustomize.toolkit.fluxcd.io/v1.
Required fields: spec.path.
The following defaults are applied to omitted field:
- metadata.name is defaulted to &ldquo;flux-system&rdquo;
- metadata.namespace is defaulted to &ldquo;flux-system&rdquo;
- spec.interval is defaulted to &ldquo;1m&rdquo;</p>
</td>
</tr>
</tbody>
</table>
<h3 id="flux.extensions.gardener.cloud/v1alpha1.Source">Source
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.extensions.gardener.cloud/v1alpha1.FluxConfig">FluxConfig</a>)
</p>
<p>
<p>Source configures how to bootstrap a Flux source object.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>template</code></br>
<em>
<a href="https://fluxcd.io/flux/components/source/api/v1/#source.toolkit.fluxcd.io/v1.GitRepository">
source.toolkit.fluxcd.io/v1.GitRepository
</a>
</em>
</td>
<td>
<p>Template is a partial GitRepository object in API version source.toolkit.fluxcd.io/v1.
Required fields: spec.ref.*, spec.url.
The following defaults are applied to omitted field:
- metadata.name is defaulted to &ldquo;flux-system&rdquo;
- metadata.namespace is defaulted to &ldquo;flux-system&rdquo;
- spec.interval is defaulted to &ldquo;1m&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>secretResourceName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecretResourceName references a resource under Shoot.spec.resources.
The secret data from this resource is used to create the GitRepository&rsquo;s credentials secret
(GitRepository.spec.secretRef.name) if specified in Template.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
