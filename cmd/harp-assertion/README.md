# Harp - Assertion

A `harp` plugin to generate and verify custom JWT Assertions to be used
as credentials for `Vault` JWT authentication backend or simply `private_jwt` client
credentials.

> Because principal/secret is not the only authentication scheme available !

## Build

```sh
export PATH=<harp-repository-path>/tools/bin:$PATH
mage
```

## Install

Stable release

```sh
brew install elastic/harp-plugins/harp-attestation
```

Built from source

```sh
brew install --from-source elastic/harp-plugins/harp-attestation
```

## Create a transit key

A Transit key is a key handled by Vault so that private key stay inside Vault
and interaction are made using Transit backend API.

First, connect to Vault.

```json
export VAULT_ADDR=<vault-url>
export VAULT_TOKEN=$(vault login -method=oidc -token-only)
```

Enable transit backend

```sh
$ vault secrets enable transit -path=assertions
Success! Enabled the transit secrets engine at: assertions/
```

Create a dedicated assertion key using P-384 curve

```sh
$ vault write -f assertions/keys/product1-ci-jobs type=ecdsa-p384
Success! Data written to: assertions/keys/product1-ci-jobs
```

## Generate an assertion

Prepare a JSON body, this JSON body will be integrated and protected in the
assertion by cryptographic signature.

```json
{
    "project_id": "123456789",
    "ref": "master",
    "ref_type": "branch",
    "git_commit_ref": "ae15fda456",
    "roles": ["ci:docker", "ci:pusher"],
    "project_path": "mygroup/myproject",
    "user_id": "42",
    "user_login": "myuser",
    "user_email": "myuser@example.com",
    "pipeline_id": "1212",
    "job_id": "1212"
}
```

Use `harp-assertion` to seal the JSON body with the Vault transit key nammed `product1-ci-jobs`.

```sh
export VAULT_TOKEN=$(vault write -field=token auth/approle/login role_id=$JENKINS_ROLE_ID secret_id=$JENKINS_SECRET_ID)
export CI_JOB_JWT=$(harp-assertion --in body.json --key product1-ci-jobs --sub $JENKINS_JOB_ID --audience="https://vault.domain.tld")
```

This `assertion` must be passed to job and could act as an authentication token.

## Authentication using assertion

Assertion authentication is based on proof of validity based on token signature
validation and claims value authorizations.

> So that anyone that have to public key and trust it could validate the assertion,
> at a small cost (ecdsa signature validation).

Enable `jwt` authentication backend

```sh
$ vault auth enable jwt
Success! Enabled jwt auth method at: jwt/
```

Create policies

```sh
$ vault policy write product1-staging - <<EOF
# Policy name: product1-staging
#
# Read-only permission on 'app/staging/security/cluster/v1.0.0/product1/vendors/*' path
path "app/data/staging/security/cluster/v1.0.0/product1/vendors/*" {
  capabilities = [ "read" ]
}
EOF
Success! Uploaded policy: product1-staging

$ vault policy write product1-production - <<EOF
# Policy name: product1-production
#
# Read-only permission on 'app/production/security/cluster/v1.0.0/product1/vendors/*' path
path "app/data/production/security/cluster/v1.0.0/product1/vendors/*" {
  capabilities = [ "read" ]
}
EOF
Success! Uploaded policy: product1-production
```

And then 2 roles on `jwt` backend that match policies

> For more information about role creation parmaeters - <https://www.vaultproject.io/api/auth/jwt#create-role>

A `product1-staging` role

```sh
$ vault write auth/jwt/role/product1-staging - <<EOF
{
  "role_type": "jwt",
  "bound_audiences": ["https://vault.domain.tld"],
  "policies": ["product1-staging"],
  "token_explicit_max_ttl": 60,
  "user_claim": "sub",
  "bound_claims": {
    "project_id": "22",
    "ref": "master",
    "ref_type": "branch"
  }
}
EOF
```

A `product1-production` role

```sh
$ vault write auth/jwt/role/product1-production - <<EOF
{
  "role_type": "jwt",
  "bound_audiences": ["https://vault.domain.tld"],
  "policies": ["product1-production"],
  "token_explicit_max_ttl": 60,
  "user_claim": "sub",
  "bound_claims_type": "glob",
  "bound_claims": {
    "project_id": "22",
    "ref_protected": "true",
    "ref_type": "branch",
    "ref": "auto-deploy-*"
  }
}
EOF
```

> Try to be the most restrictive as possible during `jwt` role declaration.

## Setup Public keys

### Static

> Don't forget to republish public keys in case on rotation.

Enable plublic key validation :

> Only generated assertions will be validated by Vault

```sh
harp-assertion --key jenkins --pem > key.pub
vault write auth/jwt/config \
    jwt_validation_pubkeys=@key.pub \
    bound_issuer="harp-assertion"
```

### Dynamic

Now setup public keyset on Vault backend to validate assertion.

Export JWKS

```sh
harp-assertion jwks --key product1-ci-jobs --out product1-ci-jobs.json
mv product1-ci-jobs.json <webserver>/htdocs
```

> Expose public keyset using an classic HTTP server.

```sh
$ vault write auth/jwt/config \
    jwks_url="https://internal.domain.tld/product1-ci-jobs.json" \
    bound_issuer="harp-assertion"
```

JWKS must be exposed and accessible from Vault server, so that in case of rotation
the file could be pulled to validate new assertions signed by the newest key.

## Login with assertion

```sh
export VAULT_TOKEN="$(vault write -field=token auth/jwt/login role=product1-production jwt=$CI_JOB_JWT)"
```

Use your `VAULT_TOKEN` as usual

```sh
export DOCKER_REGISTRY_TOKEN=$(vault read -field token docker-registry/roles/product1)
echo '{"auths":{"registry-1.docker.io":{"registrytoken": "$DOCKER_REGISTRY_TOKEN"}}}' | jq -s ".[0] * .[1]" ~/.docker/config.json - > ~/.docker/config.json
...
docker push registry-1.docker.io/org/product1
```

## Validate and extract payload

You can extract sealed data from assertion, if you need to use it inside your scripts.

```sh
$ wget https://internal.domain.tld/product1-ci-jobs.json
$ echo $CI_JWT_JOB | harp-attestation verify --jwks product1-ci-jobs.json
{
    "project_id": "123456789",
    "ref": "master",
    "ref_type": "branch",
    "git_commit_ref": "ae15fda456",
    "roles": ["ci:docker", "ci:pusher"],
    "project_path": "mygroup/myproject",
    "user_id": "42",
    "user_login": "myuser",
    "user_email": "myuser@example.com",
    "pipeline_id": "1212",
    "job_id": "1212"
}
```

## Conclusion

Assertion is a way to authenticate a bunch of data, and use this assertion as
an authentication token for Vault, but the same assertion could be used to
authenticate to any service that is compatiblie with `private_jwt` client
credentials.

This for example the way how GCP Service Account accesses are authenticated,
the JSON file you have is the private key used to generate a JWT assertion.
The server only need to know the public key to validate the assertion container,
no more long and cpu-intensive hash functions used to store credentials.
