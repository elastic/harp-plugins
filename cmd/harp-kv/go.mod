module github.com/elastic/harp-plugins/cmd/harp-kv

go 1.16

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/coreos/etcd v2.3.8+incompatible // indirect
	github.com/elastic/harp v0.1.20
	github.com/fatih/color v1.12.0
	github.com/hashicorp/consul/api v1.1.0
	github.com/magefile/mage v1.11.0
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/spf13/cobra v1.2.1
	go.etcd.io/etcd v2.3.8+incompatible // indirect
	go.etcd.io/etcd/client/v2 v2.305.0 // indirect
	go.etcd.io/etcd/client/v3 v3.5.0
	go.uber.org/zap v1.18.1
)
