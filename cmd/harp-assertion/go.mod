module github.com/elastic/harp-plugins/cmd/harp-assertion

go 1.15

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/elastic/harp v0.1.12
	github.com/fatih/color v1.10.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/magefile/mage v1.11.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/spf13/cobra v1.1.3
	go.uber.org/zap v1.16.0
	gopkg.in/square/go-jose.v2 v2.5.1
)
