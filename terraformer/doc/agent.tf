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
