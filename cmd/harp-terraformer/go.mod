module github.com/elastic/harp-plugins/cmd/harp-terraformer

go 1.16

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/elastic/harp v0.1.19
	github.com/fatih/color v1.12.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/golang/protobuf v1.5.2
	github.com/google/gofuzz v1.2.0
	github.com/gosimple/slug v1.9.0
	github.com/hashicorp/hcl/v2 v2.10.1
	github.com/magefile/mage v1.11.0
	github.com/spf13/cobra v1.2.1
	go.uber.org/zap v1.18.1
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)
