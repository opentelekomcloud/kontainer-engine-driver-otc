module github.com/opentelekomcloud/kontainer-engine-driver-otc

go 1.14

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/client-go => k8s.io/client-go v0.18.0
)

require (
	github.com/getlantern/deepcopy v0.0.0-20160317154340-7f45deb8130a
	github.com/opentelekomcloud-infra/crutch-house v0.2.4-0.20201120120921-faec6632f2ed
	github.com/opentelekomcloud/gophertelekomcloud v0.1.1-0.20201120102720-b72395887513
	github.com/rancher/kontainer-engine v0.0.4-dev.0.20200406202044-bf3f55d3710a
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	k8s.io/client-go v12.0.0+incompatible
)
