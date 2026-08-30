# GitHub Permissions Feature

## Overview

The GitHub Permissions feature allows you to define GitHub App permission sets that will be stored in Vault as generic secrets. These permission sets can then be used by your applications to create GitHub App installations with specific repository access and permissions.

## Configuration

Add a `githubPermissionSets` (or `github_permission_sets` in snake_case) section to your AppRoleDefinition spec. 

```yaml
spec:
  githubPermissionSets:
    - name: "permission-set-name"
      installation_id: "local.github_installation_id"
      org_name: "local.github_organization_name"
      repositories: [1057520243, 1057520244]
      permissions:
        metadata: "read"
        contents: "write"
        pull_requests: "write"
```

### Fields

- **name** (required): A unique identifier for the permission set. This will be used in the Vault path and Terraform resource name.
- **installation_id** (required): Reference to the GitHub App installation ID. Typically passed as a local reference (e.g., `local.github_installation_id`).
- **org_name** (required): Reference to the GitHub organization name. Typically passed as a local reference (e.g., `local.github_organization_name`).
- **repositories** (required): Array of repository IDs that this permission set applies to.
- **permissions** (required): Map of GitHub permission names to permission levels (e.g., "read", "write", "admin").

## Generated Terraform

For each GitHub permission set, the compiler generates a `vault_generic_secret` resource:

```hcl
resource "vault_generic_secret" "github-permissionset-<name>-<environment>" {
  path = "${local.github_secrets_engine_path}/permissionset/<name>-<environment>"

  data_json = jsonencode({
    installation_id = <installation_id_reference>
    org_name        = <org_name_reference>

    permissions = {
      metadata      = "read"
      contents      = "write"
      pull_requests = "write"
    }

    repository_ids = [
      1057520243,
      1057520244
    ]
    repositories = null
  })
}
```

## Example

See [github-permissions-example.yaml](./github-permissions-example.yaml) for a complete example configuration.

## Usage

1. Define your GitHub permission sets in the AppRoleDefinition YAML
2. Run the terraformer compiler to generate Terraform files
3. The generated Terraform will create Vault secrets containing the permission configurations
4. Your applications can read these secrets from Vault to configure GitHub App installations

## Notes

- The `repositories` field is set to `null` in the generated Terraform to prioritize `repository_ids`
- Permission set names are automatically suffixed with the environment (e.g., `dev`, `staging`, `production`)
- The installation_id and org_name fields support Terraform local references for dynamic configuration
