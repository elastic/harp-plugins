module github.com/elastic/harp-plugins/cmd/harp-server

go 1.16

// Snyk finding
replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	cloud.google.com/go/storage v1.16.0
	github.com/Azure/azure-sdk-for-go v55.8.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.19 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/awnumar/memguard v0.22.2
	github.com/aws/aws-sdk-go v1.40.7
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/dnaeon/go-vcr v1.2.0 // indirect
	github.com/elastic/harp v0.1.19
	github.com/fatih/color v1.12.0
	github.com/go-chi/chi v1.5.4
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/golang/mock v1.6.0
	github.com/google/wire v0.5.0
	github.com/gosimple/slug v1.9.0
	github.com/hashicorp/vault/api v1.1.1
	github.com/magefile/mage v1.11.0
	github.com/oklog/run v1.1.0
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.2.1
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.39.0
)
