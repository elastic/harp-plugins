# Harp Linter

`harp` plugin that allows you to :

* Lint a `Bundle` content

## Build

```sh
export PATH=<harp-repository-path>/tools/bin:$PATH
mage
```

## Install

Stable release

```sh
brew install elastic/harp-plugins/harp-linter
```

Built from source

```sh
brew install --from-source elastic/harp-plugins/harp-linter
```

## Constraint Language

In order to produce package constraints, `harp-linter` uses [CEL](https://github.com/google/cel-go).

For complete language specification, consult this repository - <https://github.com/google/cel-spec>

### Extensions

* `p` exposes the current package that match the `path` filter;
* `p.match_path(string)` return true or false if current package match the given [glob](https://github.com/gobwas/glob) path filter;
* `p.is_cso_compliant()` return true or false according to CSO Compliance state of the current package;
* `p.has_secret(string)` return true or false according to secret key `string` existence;
* `p.has_all_secrets(list)` return true or false if package has all given secret keys;

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
        - p.match_path("app/*")
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
