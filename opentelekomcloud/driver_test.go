package opentelekomcloud

import (
	"context"
	"os"
	"testing"

	"github.com/opentelekomcloud-infra/crutch-house/utils"
	"github.com/sirupsen/logrus"

	"github.com/opentelekomcloud-infra/crutch-house/services"
	"github.com/rancher/kontainer-engine/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	charset           = "0123456789abcdefghijklmnopqrstuvwxyz"
	authFailedMessage = "failed to authorize client"
	defaultNodeCount  = 2
)

var (
	kpName              = utils.RandomString(10, "kp-")
	vpcName             = utils.RandomString(10, "vpc-")
	subnetName          = utils.RandomString(10, "subnet-")
	kontainerDriverName = utils.RandomString(10, "kd-")
	name                = utils.RandomString(10, "c-", charset)
)

func getDriverOpts() *types.DriverOptions {
	stringOptions := map[string]string{
		"accessKey":            os.Getenv("accessKey"),
		"authenticationMode":   "rbac",
		"availabilityZone":     "eu-de-03",
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
		"name":                 name,
		"nodeFlavor":           "s2.large.2",
		"nodeOs":               "EulerOS 2.5",
		"password":             os.Getenv("password"),
		"projectName":          os.Getenv("projectName"),
		"region":               "eu-de",
		"rootVolumeType":       "SATA",
		"secretKey":            os.Getenv("secretKey"),
		"subnet":               subnetName,
		"username":             os.Getenv("username"),
		"vpc":                  vpcName,
		"appProtocol":          "TCP",
	}
	intOptions := map[string]int64{
		"clusterEipBandwidthSize": 10,
		"dataVolumeSize":          100,
		"rootVolumeSize":          40,
		"nodeCount":               defaultNodeCount,
		"appPort":                 80,
	}
	stringSliceOptions := map[string]*types.StringSlice{
		"clusterLabels": {
			Value: []string{"origin=rancher-otc"},
		},
	}

	driverOptions := types.DriverOptions{
		StringOptions:      stringOptions,
		IntOptions:         intOptions,
		StringSliceOptions: stringSliceOptions,
	}

	return &driverOptions
}

func GetNewIntOpts() map[string]int64 {
	intOptions := map[string]int64{
		"clusterEipBandwidthSize": 10,
		"dataVolumeSize":          100,
		"rootVolumeSize":          40,
		"nodeCount":               defaultNodeCount + 1,
		"appPort":                 80,
	}
	return intOptions
}

func authClient(t *testing.T) services.Client {
	client, err := services.NewClient("OS_")
	require.NoError(t, err, "failed to initialize client")
	err = client.Authenticate()
	require.NoError(t, err, authFailedMessage)
	return client
}

func computeClient(t *testing.T) services.Client {
	client := authClient(t)
	require.NoError(t, client.InitCompute())
	return client
}

func TestDriver_ClusterWorkflow(t *testing.T) {
	driverOptions := getDriverOpts()

	ctx := context.Background()

	client := computeClient(t)
	_, err := client.CreateKeyPair(kpName, "")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, client.DeleteKeyPair(kpName))
	}()

	driver := NewDriver()
	info, err := driver.Create(ctx, driverOptions, &types.ClusterInfo{})
	require.NoError(t, err)

	info, err = driver.PostCheck(ctx, info)
	assert.NoError(t, err)

	newDriverOptions := getDriverOpts()
	newDriverOptions.IntOptions = GetNewIntOpts()

	logrus.Info("Update cluster by adding 1 node")
	info, err = driver.Update(ctx, info, newDriverOptions)
	assert.NoError(t, err)

	logrus.Info("Update cluster by decreasing 2 nodes")
	newDriverOptions.IntOptions["nodeCount"] = defaultNodeCount - 1
	info, err = driver.Update(ctx, info, newDriverOptions)
	assert.NoError(t, err)

	err = driver.Remove(ctx, info)
	assert.NoError(t, err)
}
