# GitHub Permission Sets - Separate File Generation

## Overview

The GitHub permission sets can now be generated in a separate Terraform file using the new `github` command. This allows you to organize your Terraform code better by separating concerns.

## Usage

### Generate GitHub Permission Sets Only

To generate only the GitHub permission sets in a separate file:

```bash
harp-terraformer github --spec <your-spec.yaml> --env <environment> --out github-permissions.tf
```

Example:
```bash
harp-terraformer github --spec doc/github-permissions-example.yaml --env staging --out github-permissions.tf
```

### Generate Service Policy (without GitHub Permission Sets)

The `service` command now generates Vault policies and AppRole configurations WITHOUT GitHub permission sets:

```bash
harp-terraformer service --spec <your-spec.yaml> --env <environment> --out service.tf
```

Example:
```bash
harp-terraformer service --spec doc/github-permissions-example.yaml --env staging --out service.tf
```

### Complete Workflow

For a complete setup with both service policies and GitHub permissions:

```bash
# Generate service Terraform (policies and AppRole)
harp-terraformer service --spec your-spec.yaml --env production --out service.tf

# Generate GitHub permission sets in a separate file
harp-terraformer github --spec your-spec.yaml --env production --out github-permissions.tf
```

## Available Commands

- **`agent`** - Generate agent policy and AppRole
- **`service`** - Generate service policy and AppRole (excludes GitHub permission sets)
- **`github`** - Generate GitHub permission sets only (new!)
- **`policy`** - Generate policy document only

## Benefits

1. **Better Organization**: Keep GitHub permissions separate from Vault policies
2. **Flexibility**: Generate only what you need
3. **Modularity**: Easier to manage and version control different aspects of your infrastructure
4. **Team Collaboration**: Different teams can own different files

## Template Variables

The `github` command uses the `GitHubPermissionSetsTemplate` which includes all the same metadata as other templates:
- SpecificationHash
- Owner
- Date
- Description
- Issues

## Example Output

See the example by running:
```bash
harp-terraformer github --spec doc/github-permissions-example.yaml --env staging
```
