# Protobuf API Generation

This directory contains the Protocol Buffer definitions for the Harp Terraformer API and tooling for generating Go code and JSON Schema from these definitions.

## Directory Structure

```
api/
├── buf.yaml          # Buf workspace configuration
├── buf.gen.yaml      # Code generation configuration
├── Taskfile.yml      # Task runner commands
├── proto/            # Proto source files
│   └── harp/
│       └── terraformer/
│           └── v1/
│               └── approle.proto
└── gen/              # Generated output
    ├── go/           # Generated Go code
    └── jsonschema/   # Generated JSON Schema files
```

## Prerequisites

Install the [Buf CLI](https://buf.build/docs/cli/installation/):

```sh
# macOS
brew install bufbuild/buf/buf

# Other platforms
# See: https://buf.build/docs/cli/installation/
```

## Generation

### Generate All Outputs

Generate both Go code and JSON Schema:

```sh
task generate
# or directly:
buf generate
```

This will:
- Clean the `gen/` directory
- Generate Go protobuf code to `gen/go/`
- Generate JSON Schema files to `gen/jsonschema/`

### Go Code Generation

Go code is generated using the remote `buf.build/protocolbuffers/go` plugin with `paths=source_relative` option. Output files are placed in `gen/go/` mirroring the proto package structure.

### JSON Schema Generation

JSON Schema is generated using the [protoschema-jsonschema](https://github.com/bufbuild/protoschema-plugins) plugin from Buf with `target=json-bundle`:

- **Draft 2020-12** compliant JSON Schema
- **camelCase** field names (standard JSON convention)
- **Self-contained** (all `$ref` dependencies resolved inline)

#### Output Files

The plugin generates fully-qualified names, and a post-processing step renames them to simple consumer-friendly names:

```
gen/jsonschema/
├── AppRoleDefinition.json
├── AppRoleDefinitionMeta.json
├── AppRoleDefinitionNamespaces.json
├── AppRoleDefinitionSecretSuffix.json
├── AppRoleDefinitionSelector.json
└── AppRoleDefinitionSpec.json
```

#### Consumer URLs

Use the simple-named files for stable consumer-facing URLs:

```
https://raw.githubusercontent.com/elastic/harp-plugins/main/terraformer/api/gen/jsonschema/AppRoleDefinition.json
```

## Available Tasks

| Task | Description |
|------|-------------|
| `task generate` | Generate Go and JSON Schema from proto files |
| `task rename-schemas` | Create simple-named JSON schema copies (called by generate) |
| `task lint` | Lint proto files for style and correctness |
| `task breaking` | Check for breaking changes against main branch |
| `task format` | Format proto files in place |

## Configuration Files

### buf.yaml

Workspace configuration defining:
- Module path (`proto/`)
- Lint rules (`STANDARD`)
- Breaking change detection (`FILE`)

### buf.gen.yaml

Generation configuration defining:
- Plugins (Go, JSON Schema)
- Output directories
- Plugin options:
  - Go: `paths=source_relative`
  - JSON Schema: `target=json-bundle` (camelCase, self-contained)

## Reference Links

- [Buf CLI Quickstart](https://buf.build/docs/cli/quickstart/)
- [buf.yaml Configuration](https://buf.build/docs/configuration/v2/buf-yaml)
- [buf.gen.yaml Configuration](https://buf.build/docs/configuration/v2/buf-gen-yaml)
- [Protoschema Plugins (JSON Schema)](https://github.com/bufbuild/protoschema-plugins)
- [Buf Lint Rules](https://buf.build/docs/lint/rules/)
- [Buf Breaking Rules](https://buf.build/docs/breaking/rules/)
