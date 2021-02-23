# Harp Plugins

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/elastic/harp-plugins/graphs/commit-activity)

This repository contains [harp](https://github.com/elastic/harp) plugins. These
plugins are SDK usage samples.

## GA

* [harp-server](https://github.com/elastic/harp-plugins/tree/main/cmd/harp-server) - Bundle server to expose a bundle via a HTTP / Vault / gRPC API.
* [harp-terraformer](https://github.com/elastic/harp-plugins/tree/main/cmd/harp-terraformer) - Use to harp template engine to render Vault AppRoles / Policies to ease multiple cluster deployment from a YAML Descriptor.

## Beta

* [harp-assertion](https://github.com/elastic/harp-plugins/tree/main/cmd/harp-assertion) - Create JWT Assertion for decentralized authentication purpose.
* [harp-aws](https://github.com/elastic/harp-plugins/tree/main/cmd/harp-aws) - AWS related container operations (Container identity via KMS, S3 publication, Cloud Secret Storage management).
* [harp-linter](https://github.com/elastic/harp-plugins/tree/main/cmd/harp-linter) - Use CEL language to add Bundle constraints.
* [harp-yubikey](https://github.com/elastic/harp-plugins/tree/main/cmd/harp-yubikey) - Container identity management via a Yubikey and retired key management feature.
