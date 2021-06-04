module github.com/elastic/harp-plugins/cmd/harp-linter

go 1.16

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/elastic/harp v0.1.17
	github.com/fatih/color v1.12.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/protobuf v1.5.2
	github.com/google/cel-go v0.7.3
	github.com/magefile/mage v1.11.0
	github.com/spf13/cobra v1.1.3
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	google.golang.org/genproto v0.0.0-20210604141403-392c879c8b08
	google.golang.org/protobuf v1.26.0
	sigs.k8s.io/yaml v1.2.0
)
