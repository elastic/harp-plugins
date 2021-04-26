module github.com/elastic/harp-plugins/cmd/harp-aws

go 1.16

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/aws/aws-sdk-go v1.38.25
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/elastic/harp v0.1.15
	github.com/fatih/color v1.10.0
	github.com/magefile/mage v1.11.0
	github.com/spf13/cobra v1.1.3
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
)
