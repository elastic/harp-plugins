module github.com/elastic/harp-plugins/cmd/harp-yubikey

go 1.15

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/elastic/harp v0.1.12
	github.com/fatih/color v1.10.0
	github.com/fxamacker/cbor/v2 v2.2.0
	github.com/go-piv/piv-go v1.7.0
	github.com/magefile/mage v1.11.0
	github.com/spf13/cobra v1.1.3
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
)
