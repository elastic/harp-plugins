module github.com/elastic/harp-plugins/cmd/harp-yubikey

go 1.16

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/elastic/harp v0.1.19
	github.com/fatih/color v1.12.0
	github.com/fxamacker/cbor/v2 v2.3.0
	github.com/go-piv/piv-go v1.8.0
	github.com/magefile/mage v1.11.0
	github.com/spf13/cobra v1.2.1
	go.uber.org/zap v1.18.1
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
)
