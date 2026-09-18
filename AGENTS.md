# libsonnet-gen — Agent Notes

## What this repo does
Code generator for Jsonnet Kubernetes libraries. Consumes OpenAPI v2 (Swagger), CRDs, or JSON Schema → produces `.libsonnet` files.

## Key commands
```bash
go test ./...                          # run all tests
make test                              # same
make libs/<name>                       # generate a library from libs/<name>/config.json
make libs/<name> VERSIONS="1.27.0"     # regenerate specific versions only
go build -o libsonnet-gen .            # build binary (Makefile expects ./libsonnet-gen in repo root)
go install .                           # installs as "libsonnet-gen" (module name)
```

## Library generation
- `libs/` contains 100+ pre-configured CRD-based libraries, each with its own `config.json`
- `libsonnet-gen --target kubernetes --config libs/<name>/config.json` — run directly (add `--debug` for verbose logging)
- `libsonnet-gen --target jsonschema --schema <path> --output <path>` — generate from JSON Schema instead
- `outputDir` in config.json is relative to the config file's directory, not cwd

## Golden tests
```bash
make update-golden-test-data           # regenerate all golden .libsonnet fixtures
go test ./pkg/compiler/jsonschemacompiler/... -update   # same, targeted
```
Golden files live in `pkg/compiler/jsonschemacompiler/testdata/TestCompileLibsonnet/*.golden.libsonnet`.

## Code generation pipeline
```
config.json → specGenerator (optional) → specs[] → model.Load() → render → writer (disk)
```
- **OpenAPI/CRD path**: `pkg/model/modifiers.go` (model) → `pkg/render/modifiers.go` (render)
- **JSON Schema path**: `pkg/compiler/jsonschemacompiler/jsonschemacompiler.go` (standalone compiler)
- **Builder**: `pkg/builder/` — generates Jsonnet AST from Go (Func, Object, Call, List, etc.)
- **CLI entry**: `main.go` → `cmd/libsonnet-gen/root.go` (urfave/cli/v3)

## Modifiers model
For each field, the generator creates:
- `withXyz(value)` — replaces the value (for arrays: accepts single or list, wraps if needed)
- `withXyzMixin(value)` — appends to existing value (`+:`)
- For array-of-object fields: nested helpers like `containers: { withName(), withImage() }`
- `mapXyz(f)` — maps `std.map(f, super.field)` over each element (all arrays)
- `mapXyzByName(name, transformFunc)` — transforms only the element where `item.name == name` (arrays whose items have a `"name"` property)

## Custom patches vs extensions
- **`custom/`** — auto-merged into `main.libsonnet` during generation. Declare via `"patchDir": "custom/core"` in `config.json`.
- **`extensions/`** — NOT auto-applied. User must import manually. Declare via `"extensionDir": "extensions/core"`.

## Toolchain
- Go 1.26.6 (exact, see `go.mod`)
- `go-jsonnet` 0.22.0 for running generated libsonnet
- Managed via `mise.toml` — run `mise install` if tools are missing
- CI: `make test` only — no lint or typecheck step
- Release: push `v*` tag → GoReleaser → GitHub releases + GHCR Docker images (`.github/workflows/release.yml`)

## Gotchas
- The Makefile expects `./libsonnet-gen` in the repo root — build before running `make libs/*`.
- `specGenerator` and manual `specs` are mutually exclusive modes in `config.json`.
- GitHub auto-discovery needs `GITHUB_TOKEN` for rate limits (60/hr unauth, 5000/hr with token).
- When adding new builder types, update both `String()` output and `Name()` — the Object builder uses child `Name()` as the object key.
- Hyphenated field names (e.g. `local-prefixes`) need bracket notation in generated Jsonnet: `super['field-name']` not `super.field-name`.
