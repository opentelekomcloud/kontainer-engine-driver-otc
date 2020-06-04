package opentelekomcloud

import (
	"context"
	"os"
	"testing"

	"github.com/opentelekomcloud-infra/crutch-house/clientconfig"
	"github.com/opentelekomcloud-infra/crutch-house/services"
	"github.com/rancher/kontainer-engine/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	charset           = "0123456789abcdefghijklmnopqrstuvwxyz"
	authFailedMessage = "failed to authorize client"
)

var (
	kpName              = services.RandomString(10, "kp-")
	vpcName             = services.RandomString(10, "vpc-")
	subnetName          = services.RandomString(10, "subnet-")
	kontainerDriverName = services.RandomString(10, "kd-")
	name                = services.RandomString(10, "c-", charset)
)

func getDriverOpts() *types.DriverOptions {
	boolOptions := map[string]bool{
		"createLoadBalancer": true,
	}

	stringOptions := map[string]string{
		"accessKey":            os.Getenv("accessKey"),
		"authenticationMode":   "rbac",
		"availabilityZone":     "eu-de-01",
		"bmsPeriodType":        "month",
		"clusterEipShareType":  "PER",
		"clusterEipType":       "5_bgp",
		"clusterFlavor":        "cce.s1.small",
		"clusterType":          "VirtualMachine",
		"containerNetworkCidr": "172.16.0.0/16",
		"containerNetworkMode": "overlay_l2",
		"dataVolumeType":       "SATA",
		"description":          "test cluster",
		"domainName":           os.Getenv("domainName"),
		"driverName":           kontainerDriverName,
		"keyPair":              kpName,
		"lbEipShareType":       "PER",
		"lbEipType":            "5_bgp",
		"name":                 name,
		"nodeFlavor":           "s2.large.1",
		"nodeOs":               "EulerOS 2.5",
		"password":             os.Getenv("password"),
		"projectName":          os.Getenv("projectName"),
		"region":               "eu-de",
		"rootVolumeType":       "SATA",
		"secretKey":            os.Getenv("secretKey"),
		"subnet":               subnetName,
		"username":             os.Getenv("username"),
		"vpc":                  vpcName,
	}
	intOptions := map[string]int64{
		"clusterEipBandwidthSize": 10,
		"dataVolumeSize":          100,
		"rootVolumeSize":          40,
		"nodeCount":               1,
		"appPort":                 80,
	}
	stringSliceOptions := map[string]*types.StringSlice{
		"clusterLabels": {
			Value: []string{"origin=rancher-otc"},
		},
	}

	driverOptions := types.DriverOptions{
		BoolOptions:        boolOptions,
		StringOptions:      stringOptions,
		IntOptions:         intOptions,
		StringSliceOptions: stringSliceOptions,
	}

	return &driverOptions
}

func authClient(t *testing.T) services.Client {
	client := services.NewClient(&clientconfig.ClientOpts{
		AuthInfo: &clientconfig.AuthInfo{
			AuthURL:     authURL,
			Username:    os.Getenv("username"),
			Password:    os.Getenv("password"),
			ProjectName: os.Getenv("projectName"),
			DomainName:  os.Getenv("domainName"),
			AccessKey:   os.Getenv("accessKey"),
			SecretKey:   os.Getenv("secretKey"),
		},
		RegionName:   "eu-de",
		EndpointType: "public",
	})
	err := client.Authenticate()
	require.NoError(t, err, authFailedMessage)
	return client
}

func computeClient(t *testing.T) services.Client {
	client := authClient(t)
	require.NoError(t, client.InitCompute())
	return client
}

func TestDriver_CreateCluster(t *testing.T) {
	driverOptions := getDriverOpts()

	ctx := context.Background()

	client := computeClient(t)
	_, err := client.CreateKeyPair(kpName, "")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, client.DeleteKeyPair(kpName))
	}()

	driver := NewDriver()
	info, err := driver.Create(ctx, driverOptions, &types.ClusterInfo{})
	require.NoError(t, err)

	err = driver.Remove(ctx, info)
	logrus.Debug(err)
	require.NoError(t, err)
}
