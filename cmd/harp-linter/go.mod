module github.com/elastic/harp-plugins/cmd/harp-linter

go 1.16

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/elastic/harp v0.1.19
	github.com/fatih/color v1.12.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/protobuf v1.5.2
	github.com/google/cel-go v0.7.3
	github.com/magefile/mage v1.11.0
	github.com/spf13/cobra v1.2.1
	go.uber.org/zap v1.18.1
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	google.golang.org/genproto v0.0.0-20210722135532-667f2b7c528f
	google.golang.org/protobuf v1.27.1
	sigs.k8s.io/yaml v1.2.0
)
