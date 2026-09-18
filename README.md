# libsonnet-gen

Code generator for Jsonnet Kubernetes libraries.

This repository contains the generator code and relevant bits to generate the jsonnet
libraries. It can generate libraries directly from OpenAPI spec v2 (Swagger), from
CustomResourceDefinitions, or from JSON Schema.

## Installation

### Docker

```bash
docker pull ghcr.io/thisisibrahimd/libsonnet-gen:<version>
```

### Binary release

Download pre-built binaries from [GitHub Releases](https://github.com/thisisibrahimd/libsonnet-gen/releases):

```bash
curl -sL https://github.com/thisisibrahimd/libsonnet-gen/releases/latest/download/libsonnet-gen_linux_x86_64.tar.gz | tar xz
sudo mv libsonnet-gen /usr/local/bin/
```

### From source

```bash
go install github.com/thisisibrahimd/libsonnet-gen@latest
```

## Usage

### Create a self-managed lib

Create a directory, ideally named after the project:

```bash
mkdir libs/cloudnative-pg
```

#### Approach 1: Auto-Discovery (Recommended)

Create a `config.json` with `specGenerator` to auto-discover CRDs from GitHub:

```json
{
  "libName": "cloudnative-pg",
  "description": "Generated Jsonnet library for CloudNativePG",
  "specGenerator": {
    "type": "github",
    "repo": "https://github.com/cloudnative-pg/cloudnative-pg",
    "crdPath": "config/crd/bases",
    "prefix": "^io\\.cnpg\\.postgresql\\..*",
    "dedupeCrds": true
  },
  "specs": []
}
```

Then run:

```bash
libsonnet-gen --target kubernetes --config libs/cloudnative-pg/config.json
```

Note: Output defaults to the config file's directory. Set `outputDir` in config to change this (relative to config file).

The generator will:
1. Fetch tags from the GitHub repo
2. Parse the Git tree to find CRD files in `crdPath`
3. Skip consecutive versions whose CRD files are identical to the previous kept version
4. Generate specs for each remaining version automatically

By default, auto-discovery skips duplicate CRD versions (`dedupeCrds: true`). Set `dedupeCrds: false` in `specGenerator` to generate a library for every version.

Alternatively, pin the exact versions to generate by adding a `versions` list to `specGenerator`. `versions` and `versionLimit` are mutually exclusive; when `versions` is provided, tag discovery is skipped and `dedupeCrds` is ignored. A target is generated for every version in the list.

The `versionPrefix` field (default `"v"`) is prepended to each version in `versions` when resolving the GitHub ref/tag. Use `"versionPrefix": ""` if your versions already include the prefix (e.g. `v1.27.0`).

```json
{
  "libName": "cloudnative-pg",
  "description": "Generated Jsonnet library for CloudNativePG",
  "specGenerator": {
    "type": "github",
    "repo": "https://github.com/cloudnative-pg/cloudnative-pg",
    "crdPath": "config/crd/bases",
    "prefix": "^io\\.cnpg\\.postgresql\\..*",
    "versions": ["1.27.0", "1.30.0"]
  },
  "specs": []
}
```

**GitHub Token**: Set `GITHUB_TOKEN` environment variable to authenticate with GitHub. Without a token, you are limited to 60 requests/hour (unauthenticated rate limit). With a token, you get 5,000 requests/hour.

If you have the [GitHub CLI](https://cli.github.com/) installed, you can use it to get a token:

```bash
GITHUB_TOKEN=$(gh auth token) libsonnet-gen --target kubernetes --config libs/cloudnative-pg/config.json
```

Or export it for the session:

```bash
export GITHUB_TOKEN=$(gh auth token)
libsonnet-gen --target kubernetes --config libs/cloudnative-pg/config.json
```

#### Approach 2: Manual Specs

If you prefer to manually define versions and CRD URLs, omit `specGenerator` and populate `specs` directly:

```json
{
  "libName": "cloudnative-pg",
  "description": "Generated Jsonnet library for CloudNativePG",
  "specs": [
    {
      "output": "1.27.0",
      "prefix": "^io\\.cnpg\\.postgresql\\..*",
      "crds": [
        "https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/v1.27.0/config/crd/bases/postgresql.cnpg.io_backups.yaml",
        "https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/v1.27.0/config/crd/bases/postgresql.cnpg.io_clusters.yaml"
      ]
    },
    {
      "output": "1.30.0",
      "prefix": "^io\\.cnpg\\.postgresql\\..*",
      "crds": [
        "https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/v1.30.0/config/crd/bases/postgresql.cnpg.io_backups.yaml",
        "https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/v1.30.0/config/crd/bases/postgresql.cnpg.io_clusters.yaml"
      ]
    }
  ]
}
```

Note: When using manual specs, set a per-spec `prefix` (regex) to filter which API groups are included; without a prefix, all groups found in the specs are included. Each spec must set either `crds` (a list of CRDs) or `openapi` (a single OpenAPI v2/Swagger spec) — they are mutually exclusive, and both accept URLs or local file paths.

| Feature | Auto-Discovery | Manual |
|---------|---------------|--------|
| `specGenerator` | Required | Omitted |
| `specs` | Empty `[]` | Populated with versions/CRDs |
| `prefix` | At `specGenerator` level | Not needed |
| `versionLimit` | Default `10` (mutually exclusive with `versions`) | Not applicable |
| `versions` | Optional exact list of clean versions (without prefix) | Not applicable |
| `versionPrefix` | Default `"v"` (prepended to `versions` for the Git ref) | Not applicable |
| `includeVersions` | Default `^v?\d+\.\d+\.\d+$` (regex filter on discovered tag names) | Not applicable |
| `dedupeCrds` | Default `true` (skips duplicate CRD versions; ignored when `versions` is set) | Not applicable |
| Maintenance | Automatic on new releases | Manual updates required |

### Create or update a new lib

Create a folder in `libs/`:

```bash
mkdir libs/<name>
```

Create a `config.json` in the new folder. This example renders a lib from CRDs:

```json
{
  "libName": "<name>",
  "description": "Generated Jsonnet library for <name>",
  "specGenerator": {
    "type": "github",
    "repo": "https://github.com/<owner>/<name>",
    "crdPath": "config/crd/bases",
    "prefix": "^<prefix>\\.<name>\\..*",
    "dedupeCrds": false
  },
  "specs": []
}
```

Build the binary first (the Makefile expects `./libsonnet-gen` in the repo root):

```bash
go build -o libsonnet-gen .
```

Generate the library:

```bash
$ make libs/<name>
```

Pass `VERSIONS` to regenerate only specific versions:

```bash
$ make libs/<name> VERSIONS="1.27.0"
```

Or run the binary directly:

```bash
$ libsonnet-gen --target kubernetes --config libs/<name>/config.json
```

Append version outputs as positional arguments to regenerate only specific versions, e.g. `libsonnet-gen --target kubernetes --config libs/<name>/config.json 1.27.0`. Use the global `--debug` flag for verbose logging.

### Generate Jsonnet from JSON Schema

You can also generate a Jsonnet library directly from a JSON Schema file. This is useful for configuration files such as `.golangci.yml`, `package.json`, or any other schema-backed document.

Create a small JSON Schema. For example, `config.schema.json`:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Config",
  "type": "object",
  "properties": {
    "name": {
      "type": "string"
    },
    "enabled": {
      "type": "boolean"
    }
  }
}
```

Run the generator:

```bash
$ libsonnet-gen --target jsonschema \
  --schema ./config.schema.json \
  --output ./config.libsonnet
```

The generated `config.libsonnet` looks like:

```jsonnet
{
  withName(name): {
    name: name,
  },
  withEnabled(enabled): {
    enabled: enabled,
  },
}
```

Nested objects produce nested functions and additive objects (`+:`), so you can compose configuration pieces:

```jsonnet
// example from a schema with nested properties
{
  server: {
    withPort(port): {
      server+: {
        port: port,
      },
    },
  },
}
```

#### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--schema` | Yes | Path or URL to the JSON Schema file. |
| `--output` | No | Output file path. If omitted, prints to stdout. |

#### Array support

Array properties generate three function types:

```jsonnet
{
  // Replaces the entire array (wraps single values in a list automatically)
  withTags(tags): {
    tags: if std.isArray(v=tags) then tags else [tags],
  },
  // Appends to the existing array
  withTagsMixin(tags): {
    tags+: if std.isArray(v=tags) then tags else [tags],
  },
  // Transforms each element using a function
  mapTags(f): {
    tags: std.map(f, super.tags),
  },
}
```

For arrays of objects, nested helpers are also generated for constructing individual items:

```jsonnet
{
  containers: {
    withName(name): { containers+: { name: name } },
    withImage(image): { containers+: { image: image } },
  },
}
```

If the array items have a `"name"` field, an additional `mapXyzByName` helper is generated for targeted updates:

```jsonnet
{
  // Transforms only the element where item.name matches
  mapContainersByName(name, transformFunc): {
    containers: [if c.name == name then transformFunc(c) else c for c in super.containers],
  },
}
```

Example usage:

```jsonnet
local lib = import './config.libsonnet';

// Add containers and patch one by name
local base = lib
  + lib.withContainers([
      { name: 'app', image: 'myapp:v1' },
      { name: 'sidecar', image: 'busybox' },
    ])
  + lib.mapContainersByName('app', function(c)
      c + { image: 'myapp:v2' }
    );

// Transform all tags
base + lib.mapTags(function(t) 'release-' + t)
```

#### Other limitations

- Schema composition with `anyOf`, `oneOf`, or external `$ref` is limited.
- Empty object properties (`{}`) may appear when the schema declares an object but provides no properties.
- Nested map helpers (`mapXyz` inside a `mapXyz` transform) are not supported. The helpers use `super` which only works in object merge contexts. For nested arrays, use `std.map` directly:
  ```jsonnet
  lib.mapContainers(function(c)
    c {
      ports: std.map(function(p) p { containerPort: 8080 }, super.ports),
    }
  )
  ```


## Customizing

Because the generator only creates the most minimal yet functional code, more
sophisticated utilities like constructors (`deployment.new(name, replicas,
containers)`, etc) are not created.

For that, there are two methods for extending:

### `custom` patches

The [`custom/`](https://github.com/thisisibrahimd/libsonnet-gen/tree/master/libs/k8s/custom)
directory contains a set of `.libsonnet` files, that are _automatically merged_
with the generated result in `main.libsonnet`, so they become part of the
exported API.

For example the patches in `libs/k8s`:

```

libs/k8s/
├── config.json                      # Config to generate the k8s jsonnet libraries
└── custom
    └── core
        ├── apps.libsonnet           # Constructors for `apps/v1` (daemonSet, deployment, statefulSet), ported from `ksonnet-gen` and `kausal.libsonnet`
        ├── autoscaling.libsonnet    # Extends `autoscaling/v1` and `autoscaling/v2`
        ├── batch.libsonnet          # Constructors for `batch/v1` (cronJob), ported from `kausal.libsonnet`
        ├── core.libsonnet           # Constructors for `core/v1`, ported from `ksonnet-gen` and `kausal.libsonnet`
        ├── list.libsonnet           # Adds `core.v1.List`
        ├── rbac.libsonnet           # Adds helper functions to rbac objects
        └── volumeMounts.libsonnet   # Adds helper functions to mount volumes
```

A reference for these must also be made in the `config.json`:

```json
{
  "libName": "k8s",
  "specs": [
    {
      "output": "<version>",
      "patchDir": "custom/core"
    }
  ]
}
```

### Extensions

Extensions serve a similar purpose as `custom/` patches, but are **not
automatically applied**. However, they are still part of the final artifact, but
need to added by the user themselves.

Extensions can be applied as so:

```jsonnet
(import "github.com/jsonnet-libs/k8s-libsonnet/1.21/main.libsonnet")
+ (import "github.com/jsonnet-libs/k8s-libsonnet/extensions/<name>.libsonnet")
```

A reference for these must also be made in the `config.json`:

```json
{
  "libName": "k8s",
  "specs": [
    {
      "output": "<version>",
      "extensionDir": "extensions/core"
    }
  ]
}
```
