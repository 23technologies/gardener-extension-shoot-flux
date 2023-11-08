<p>Packages:</p>
<ul>
<li>
<a href="#flux.extensions.gardener.cloud%2fv1alpha1">flux.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="flux.extensions.gardener.cloud/v1alpha1">flux.extensions.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the Flux provider config API resources.</p>
</p>
Resource Types:
<ul></ul>
<h3 id="flux.extensions.gardener.cloud/v1alpha1.FluxConfig">FluxConfig
</h3>
<p>
<p>FluxConfig specifies how to bootstrap Flux on the shoot cluster.</p>
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
<p>Source configures how to bootstrap a Flux source object.</p>
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
<p>Kustomization configures how to bootstrap a Flux Kustomization object.</p>
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
Defaults to &ldquo;v2.1.2&rdquo;.</p>
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
