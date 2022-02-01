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
