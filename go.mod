module github.com/opentelekomcloud/kontainer-engine-driver-otc

go 1.14

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/client-go => k8s.io/client-go v0.18.0
)

require (
	github.com/huaweicloud/golangsdk v0.0.0-20200414012957-3b8a408c2816
	github.com/opentelekomcloud-infra/crutch-house v0.2.1
	github.com/rancher/kontainer-engine v0.0.4-dev.0.20200406202044-bf3f55d3710a
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	k8s.io/client-go v12.0.0+incompatible
)
