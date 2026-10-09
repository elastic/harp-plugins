# Harp - Terraformer

A `harp` plugin used to generate Terraform files for `agent` and `service` in CSO context.

## UseCases

* Generate multiple TF approles/policies based on cluster environment settings;
* Make all settings consistent and technologic stack agnostic;
* Add information links.

## Build

```sh
export PATH=<harp-repository-path>/tools/bin:$PATH
go install github.com/magefile/mage@latest

# Install required dev tools
mage tools

# Run the build
mage
```

## Sample

### Request

[embedmd]:# (doc/request.yaml)
```yaml
apiVersion: harp.elastic.co/terraformer/v1
kind: AppRoleDefinition
meta:
  name: "harp-aws-deployer"
  owner: "cloud-security@elastic.co"
  description: "Generate AWS service account"
  issues:
    - https://github.com/elastic/harp-plugins/issues/123456
    - https://github.com/elastic/harp-plugins/issues/123459
spec:
  selector:
    platform: "security"
    product: "harp"
    version: "v1.0.0"
    component: "s3-publisher"
  namespaces:
    # CSO Compliant paths
    application:
      - suffix: "containers/identities/recovery"
        description: "Container sealing recovery key"
        capabilities: ["read"]
      - suffix: "containers/identities/harp-server"
        description: "Container sealing consumer key"
        capabilities: ["read"]

  # No generated paths
  custom:
  - suffix: "{{.Values.aws.backend}}/sts/harp-deploy"
    description: "Retrieve ephemeral AWS credentials for Harp container deployment"
    capabilities: ["read"]
```

### Values

> If you have multiple clusters, you can use the template engine to render the final
> request.

Production values :

[embedmd]:# (doc/production.yaml)
```yaml
aws:
  backend: aws-123456789-production
```

Staging values :

[embedmd]:# (doc/staging.yaml)
```yaml
aws:
  backend: aws-918273654-staging
```

### Compilation

#### Agent

> Agents are trusted identity allowed to generate ephemeral Vault (secret_id)
> credentials for service

```sh
harp template --in request.yaml --values production.yaml \
  | harp terraformer agent
```

It will generate `agent` terraform :

[embedmd]:# (doc/agent.tf hcl)
```hcl
# Generated with Harp Terraformer, Don't modify.
# https://github.com/elastic/harp-plugins/tree/main/cmd/harp-terraformer
# ---
# SpecificationHash: "ofDpeXC4JswbVKJFiAB+p/6EnaM8XocmhAWPnFoJQck="
# Owner: "cloud-security@elastic.co"
# Date: "2021-02-18T08:52:03Z"
# Description: "Generate AWS service account"
# Issues:
# - https://github.com/elastic/harp-plugins/issues/123456
# - https://github.com/elastic/harp-plugins/issues/123459
# ---
#
# ------------------------------------------------------------------------------

# Create the policy
data "vault_policy_document" "agent-harp-aws-deployer-production" {

  rule {
    description  = "Allow agent to retrieve service role-id"
    path         = "auth/service/role/harp-aws-deployer-production/role-id"
    capabilities = ["read"]
  }

  rule {
    description      = "Allow agent to retrieve secret-id"
    path             = "auth/service/role/harp-aws-deployer-production/secret-id"
    capabilities     = ["create", "update"]
    min_wrapping_ttl = "1s"  # minimum allowed TTL that clients can specify for a wrapped response
    max_wrapping_ttl = "90s" # maximum allowed TTL that clients can specify for a wrapped response
  }
}

# Register the policy
resource "vault_policy" "agent-harp-aws-deployer-production" {
  name   = "agent-harp-aws-deployer-production"
  policy = data.vault_policy_document.agent-harp-aws-deployer-production.hcl
}

# ------------------------------------------------------------------------------
#
# Register the backend role
resource "vault_approle_auth_backend_role" "agent-harp-aws-deployer-production" {
  backend   = "agent"
  role_name = "harp-aws-deployer-production"

  token_policies = [
    "agent-default",
    "agent-harp-aws-deployer-production",
  ]
}
```

#### Service

> Services are concrete secret consumers

```sh
harp template --in request.yaml --values production.yaml \
  | harp terraformer service
```

[embedmd]:# (doc/service.tf hcl)
```hcl
# Generated with Harp Terraformer, Don't modify.
# https://github.com/elastic/harp-plugins/tree/main/cmd/harp-terraformer
# ---
# SpecificationHash: "ofDpeXC4JswbVKJFiAB+p/6EnaM8XocmhAWPnFoJQck="
# Owner: "cloud-security@elastic.co"
# Date: "2021-02-18T08:52:40Z"
# Description: "Generate AWS service account"
# Issues:
# - https://github.com/elastic/harp-plugins/issues/123456
# - https://github.com/elastic/harp-plugins/issues/123459
# ---
#
# ------------------------------------------------------------------------------

# Create the policy
data "vault_policy_document" "service-harp-aws-deployer-production" {
  # Application secrets
  rule {
    description  = "Container sealing recovery key"
    path         = "app/data/production/security/harp/v1.0.0/s3-publisher/containers/identities/recovery"
    capabilities = ["read"]
  }

  rule {
    description  = "Container sealing consumer key"
    path         = "app/data/production/security/harp/v1.0.0/s3-publisher/containers/identities/harp-server"
    capabilities = ["read"]
  }

  # Custom secret paths
  rule {
    description  = "Retrieve ephemeral AWS credentials for Harp container deployment"
    path         = "aws-123456789-production/sts/harp-deploy"
    capabilities = ["read"]
  }
}

# Register the policy
resource "vault_policy" "service-harp-aws-deployer-production" {
  name   = "service-harp-aws-deployer-production"
  policy = data.vault_policy_document.service-harp-aws-deployer-production.hcl
}

# ------------------------------------------------------------------------------
#
# Register the backend role
resource "vault_approle_auth_backend_role" "harp-aws-deployer-production" {
  backend   = "service"
  role_name = "harp-aws-deployer-production"

  token_policies = [
    "service-default",
    "service-harp-aws-deployer-production",
  ]
}
```

## Provenance

Each generated file starts with a header that ties it back to its input.

| Header line | Meaning |
|---|---|
| `SpecificationHash` | base64 sha256 of the parsed specification (its protobuf serialization). It cannot be recomputed from the YAML file. |
| `SourceSHA256` | hex sha256 of the raw specification bytes. Always present, including for stdin. |
| `GitRepo` | Repository name, for example `elastic/harp-plugins`. Omitted when unknown. |
| `SourceFile` | Repo-relative path of the specification. Omitted when unknown. |

Check the specification against `SourceSHA256` with `sha256sum`:

```sh
sha256sum request.yaml
```

The hash covers the bytes that `harp-terraformer` reads. If you pipe the
specification through another tool, such as `harp template`, hash that tool's
output instead of the original file.

`GitRepo` and `SourceFile` are read from git when `--spec` is a file inside a
repository. The repository name comes from the `origin` remote, with any
credentials removed. When the specification comes from stdin, or git is not
available, they are omitted and a warning is logged where git is the cause.

Set them explicitly with flags, which take priority over git for each field:

```sh
gh api repos/elastic/example/contents/specs/a.yaml --jq .content | base64 -d \
  | harp-terraformer service --spec - \
      --source-repo elastic/example --source-file specs/a.yaml
```

The values are recorded as given, so they are only as trustworthy as the caller.

To fail instead of omitting the fields, for example in CI, add
`--require-provenance`. The command exits with an error that names the reason:
the specification is read from stdin, git is not installed, the path is not in a
git repository, or the repo or file is still unknown.

| Flag | Applies to | Default |
|---|---|---|
| `--source-repo` | `service`, `agent`, `policy` | none |
| `--source-file` | `service`, `agent`, `policy` | none |
| `--require-provenance` | `service`, `agent`, `policy` | off |
