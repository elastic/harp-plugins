# Harp Linter

`harp` plugin that allows you to :

* Lint a `Bundle` content

## Build

```sh
export PATH=<harp-repository-path>/tools/bin:$PATH
mage
```

## Constraint Language

In order to produce package constraints, `harp-linter` uses [CEL](https://github.com/google/cel-go).

For complete language specification, consult this repository - <https://github.com/google/cel-spec>

### Extensions

#### Package

* `p` exposes the current package that match the `path` filter;
* `p.match_path(string)` return true or false if current package match the given [glob](https://github.com/gobwas/glob) path filter;
* `p.is_cso_compliant()` return true or false according to CSO Compliance state of the current package;
* `p.has_secret(string)` return true or false according to secret key `string` existence;
* `p.has_all_secrets(list)` return true or false if package has all given secret keys;

#### Secret

* `p.secret(string)` to retrieve the named secret from the given package
* `p.secret(string).is_base64()` check if value is a valid Base64 string
* `p.secret(string).is_required()` check if value is not empty / blank
* `p.secret(string).is_url()` check if value is a valid URL string
* `p.secret(string).is_uuid()` check if value is a valid UUID string
* `p.secret(string).is_email()` check if value is a valid Email string
* `p.secret(string).is_json()` check if value is a valid JSON string

### RuleSet

[Definition](api/proto/harp/linter/v1/linter.proto)

```yaml
apiVersion: harp.elastic.co/linter/v1
kind: RuleSet
meta:
  name: harp-server
  description: Package and secret constraints for harp-server
  owner: security@elastic.co
spec:
  rules:
    # Rule identifier used to violation report
    - name: HARP-SRV-0001
      # Human readable definition of the rule.
      description: All package paths must be CSO compliant
      # Package path matcher.
      path: "*"
      # CEL constraints expressions (implicit AND between all contraints)
      constraints:
        - p.is_cso_compliant()

    # Rule identifier used to violation report
    - name: HARP-SRV-0002
      # Human readable definition of the rule.
      description: Production application should have a JWK defined
      # Package path matcher.
      path: "app/production/**/oidc"
      # CEL constraints expressions (implicit AND between all contraints)
      constraints:
        - p.has_secret("jwk") && p.secret("jwk").is_base64()
```

## Sample

### Check is all packages are CSO compliant

```yaml
apiVersion: harp.elastic.co/linter/v1
kind: RuleSet
meta:
  name: harp-server
  description: Package and secret constraints for harp-server
  owner: security@elastic.co
spec:
  rules:
    - name: HARP-SRV-0001
      description: All package paths must be CSO compliant
      path: "*"
      constraints:
        - p.is_cso_compliant()
```

Lint an empty bundle will raise an error.

```sh
$ echo '{}' | harp from jsonmap \
  | harp-linter bundle lint --spec test/fixtures/ruleset/valid/cso.yaml
{"level":"fatal","@timestamp":"2021-02-23T10:24:45.852Z","@caller":"cobra@v1.1.3/command.go:856","@message":"unable to execute task","@appName":"harp-bundle-lint","@version":"","@revision":"8ebf40d","@appID":"BfGZbI8QYmSaXsBMWj8j0EASE67QcoP4OnC8nLl8xSXXtsY3PFEaABdfvm6c9yb3","@fields":{"error":"unable to validate given bundle: rule 'HARP-SRV-0001' didn't match any packages"}}
```

Lint valid bundle

```sh
$ echo '{"infra/aws/security/eu-central-1/ec2/ssh/default/authorized_keys":{"admin":"..."}}' \
  | harp from jsonmap \
  | harp-linter bundle lint --spec test/fixtures/ruleset/valid/cso.yaml
```

> No output and exit code (0) when everything is ok

### Validate a secret structure

```yaml
apiVersion: harp.elastic.co/linter/v1
kind: RuleSet
meta:
  name: harp-server
  description: Package and secret constraints for harp-server
  owner: security@elastic.co
spec:
  rules:
    - name: HARP-SRV-0002
      description: Database credentials
      path: "app/qa/security/harp/v1.0.0/server/database/credentials"
      constraints:
        - p.has_all_secrets(['DB_HOST','DB_NAME','DB_USER','DB_PASSWORD'])
```

Lint an empty bundle will raise an error.

```sh
$ echo '{}' | harp from jsonmap \
  | harp-linter bundle lint --spec test/fixtures/ruleset/valid/database-secret-validator.yaml
{"level":"fatal","@timestamp":"2021-02-23T10:31:05.792Z","@caller":"cobra@v1.1.3/command.go:856","@message":"unable to execute task","@appName":"harp-bundle-lint","@version":"","@revision":"8ebf40d","@appID":"2kl6OWqgNTHkBumvlEtelxpJ4V1uDQCtE5MlOS1hXaUbOYtU1rrXbEL2zswx65y4","@fields":{"error":"unable to validate given bundle: rule 'HARP-SRV-0002' didn't match any packages"}}
```

Lint an invalid bundle

```sh
echo '{"app/qa/security/harp/v1.0.0/server/database/credentials":{}}' \
  | harp from jsonmap \
  | harp-linter bundle lint --spec test/fixtures/ruleset/valid/database-secret-validator.yaml
{"level":"fatal","@timestamp":"2021-02-23T10:31:24.287Z","@caller":"cobra@v1.1.3/command.go:856","@message":"unable to execute task","@appName":"harp-bundle-lint","@version":"","@revision":"8ebf40d","@appID":"7pflS7bCAAsDcAiPJWm36pypWY3nHhqOQwCc9Vp1ABCm8ZUWbmGinGL5zbP1EWvn","@fields":{"error":"unable to validate given bundle: package 'app/qa/security/harp/v1.0.0/server/database/credentials' doesn't validate rule 'HARP-SRV-0002'"}}
```

### Generate a ruleset from a bundle

It will use the input bundle structure to generate a `RuleSet`.

```sh
harp-linter ruleset from-bundle --in customer.bundle
```

```yaml
api_version: harp.elastic.co/linter/v1
kind: RuleSet
meta:
  description: Generated from bundle content
  name: vjz70BPFJuQhm_7quRGNt1ybocQU6DeXCn8h1o4aPm80CI4pM8lNwVBTDqH8SpW0W1r-8dXSVQK67pO-vtgS_Q
spec:
  rules:
  - constraints:
    - p.has_secret("API_KEY")
    name: LINT-vjz70B-1
    path: app/production/customer1/ece/v1.0.0/adminconsole/authentication/otp/okta_api_key
  - constraints:
    - p.has_secret("host")
    - p.has_secret("port")
    - p.has_secret("options")
    - p.has_secret("username")
    - p.has_secret("password")
    - p.has_secret("dbname")
    name: LINT-vjz70B-2
    path: app/production/customer1/ece/v1.0.0/adminconsole/database/usage_credentials
  - constraints:
    - p.has_secret("cookieEncryptionKey")
    - p.has_secret("sessionSaltSeed")
    - p.has_secret("jwtHmacKey")
    name: LINT-vjz70B-3
    path: app/production/customer1/ece/v1.0.0/adminconsole/http/session
  - constraints:
    - p.has_secret("API_KEY")
    name: LINT-vjz70B-4
    path: app/production/customer1/ece/v1.0.0/adminconsole/mailing/sender/mailgun_api_key
  - constraints:
    - p.has_secret("emailHashPepperSeedKey")
    name: LINT-vjz70B-5
    path: app/production/customer1/ece/v1.0.0/adminconsole/privacy/anonymizer
  - constraints:
    - p.has_secret("host")
    - p.has_secret("port")
    - p.has_secret("options")
    - p.has_secret("username")
    - p.has_secret("password")
    - p.has_secret("dbname")
    name: LINT-vjz70B-6
    path: app/production/customer1/ece/v1.0.0/userconsole/database/usage_credentials
  - constraints:
    - p.has_secret("privateKey")
    - p.has_secret("publicKey")
    name: LINT-vjz70B-7
    path: app/production/customer1/ece/v1.0.0/userconsole/http/certificate
  - constraints:
    - p.has_secret("cookieEncryptionKey")
    - p.has_secret("sessionSaltSeed")
    - p.has_secret("jwtHmacKey")
    name: LINT-vjz70B-8
    path: app/production/customer1/ece/v1.0.0/userconsole/http/session
  - constraints:
    - p.has_secret("user")
    - p.has_secret("password")
    name: LINT-vjz70B-9
    path: infra/aws/essp-customer1/us-east-1/rds/adminconsole/accounts/root_credentials
  - constraints:
    - p.has_secret("API_KEY")
    - p.has_secret("ca.pem")
    name: LINT-vjz70B-10
    path: platform/production/customer1/us-east-1/billing/recurly/vendor_api_key
  - constraints:
    - p.has_secret("username")
    - p.has_secret("password")
    name: LINT-vjz70B-11
    path: platform/production/customer1/us-east-1/postgresql/admiconsole/admin_credentials
  - constraints:
    - p.has_secret("username")
    - p.has_secret("password")
    name: LINT-vjz70B-12
    path: platform/production/customer1/us-east-1/zookeeper/accounts/admin_credentials
  - constraints:
    - p.has_secret("privateKey")
    - p.has_secret("publicKey")
    name: LINT-vjz70B-13
    path: product/ece/v1.0.0/artifact/signature/key
```
