module github.com/opentelekomcloud/kontainer-engine-driver-otc

go 1.14

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/client-go => k8s.io/client-go v0.18.0
)

require (
	github.com/getlantern/deepcopy v0.0.0-20160317154340-7f45deb8130a
	github.com/opentelekomcloud-infra/crutch-house v0.2.4-0.20201006151002-76881d602a1b
	github.com/opentelekomcloud/gophertelekomcloud v0.0.9-0.20201006135536-46b58478d4ea
	github.com/rancher/kontainer-engine v0.0.4-dev.0.20200406202044-bf3f55d3710a
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.6.1
	k8s.io/client-go v12.0.0+incompatible
)
