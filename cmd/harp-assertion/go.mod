module github.com/elastic/harp-plugins/cmd/harp-assertion

go 1.16

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/elastic/harp v0.1.19
	github.com/fatih/color v1.12.0
	github.com/hashicorp/vault/api v1.1.1
	github.com/magefile/mage v1.11.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/spf13/cobra v1.2.1
	go.uber.org/zap v1.18.1
	gopkg.in/square/go-jose.v2 v2.6.0
)
