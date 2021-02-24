module github.com/elastic/harp-plugins/cmd/harp-linter

go 1.15

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/elastic/harp v0.1.12
	github.com/fatih/color v1.10.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/protobuf v1.4.3
	github.com/google/cel-go v0.7.2
	github.com/magefile/mage v1.11.0
	github.com/spf13/cobra v1.1.3
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	google.golang.org/genproto v0.0.0-20210203152818-3206188e46ba
	google.golang.org/protobuf v1.25.0
	sigs.k8s.io/yaml v1.2.0
)
