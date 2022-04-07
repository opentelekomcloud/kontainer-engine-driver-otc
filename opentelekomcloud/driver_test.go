package opentelekomcloud

import (
	"context"
	"testing"

	"github.com/opentelekomcloud-infra/crutch-house/services"
	"github.com/opentelekomcloud-infra/crutch-house/utils"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodes"
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
	osEnv               = openstack.NewEnv("OS_")
)

func getDriverOpts(t *testing.T) *types.DriverOptions {
	cloud, err := osEnv.Cloud()
	require.NoError(t, err)
	opts, err := openstack.AuthOptionsFromInfo(&cloud.AuthInfo, cloud.AuthType)
	require.NoError(t, err)
	canonicalOpts, ok := opts.(golangsdk.AuthOptions)
	require.True(t, ok, "Incorrect auth options provided")
	stringOptions := map[string]string{
		"token":                canonicalOpts.TokenID,
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
		"domainName":           canonicalOpts.DomainName,
		"driverName":           kontainerDriverName,
		"keyPair":              kpName,
		"name":                 name,
		"nodeFlavor":           "s2.large.2",
		"nodeOs":               "EulerOS 2.5",
		"password":             canonicalOpts.Password,
		"projectName":          canonicalOpts.TenantName,
		"region":               "eu-de",
		"rootVolumeType":       "SATA",
		"subnet":               subnetName,
		"username":             canonicalOpts.Username,
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
		"nodeCount":               defaultNodeCount,
		"appPort":                 80,
	}
	return intOptions
}

func authClient(t *testing.T) *services.Client {
	client, err := services.NewClient("OS_")
	require.NoError(t, err, "failed to initialize client")
	err = client.Authenticate()
	require.NoError(t, err, authFailedMessage)
	return client
}

func computeClient(t *testing.T) *services.Client {
	client := authClient(t)
	require.NoError(t, client.InitCompute())
	return client
}

func getRealClusterState(t *testing.T, info *types.ClusterInfo) clusterState {
	state, err := infoToState(info)
	if err != nil {
		t.Errorf("error getting cluster info: %s", err)
	}
	client, err := getClient(state)
	if err != nil {
		t.Errorf("error creating CCE client: %s", err)
	}

	cluster, err := client.GetCluster(state.ClusterID)
	if err != nil {
		t.Errorf("error retrieving cluster info: %s", err)
	}

	cceClient, err := client.NewServiceClient("cce")
	if err != nil {
		t.Errorf("error creating new CCE client: %s", err)
	}
	nodeList, err := nodes.List(cceClient, cluster.Metadata.Id, nodes.ListOpts{})
	if err != nil {
		t.Errorf("error listing CCE nodes: %s", err)
	}
	nodeIDs := make([]string, len(nodeList))
	for i, node := range nodeList {
		nodeIDs[i] = node.Metadata.Id
	}

	return clusterState{
		ClusterID:            state.ClusterID,
		ClusterType:          cluster.Spec.Type,
		ClusterFlavor:        cluster.Spec.Flavor,
		ContainerNetworkMode: cluster.Spec.ContainerNetwork.Mode,
		ContainerNetworkCidr: cluster.Spec.ContainerNetwork.Cidr,
		VpcID:                cluster.Spec.HostNetwork.VpcId,
		SubnetID:             cluster.Spec.HostNetwork.SubnetId,
		HighwaySubnetID:      cluster.Spec.HostNetwork.HighwaySubnet,
		ClusterFloatingIP:    cluster.Spec.PublicIP,
		NodeConfig:           services.CreateNodesOpts{},
		NodeIDs:              nodeIDs,
	}
}

func TestDriver_ClusterWorkflow(t *testing.T) {
	driverOptions := getDriverOpts(t)

	ctx := context.Background()

	client := computeClient(t)
	_, err := client.CreateKeyPair(kpName, "")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, client.DeleteKeyPair(kpName))
	}()

	driver := NewDriver()
	info, err := driver.Create(ctx, driverOptions, nil)
	require.NoError(t, err)
	t.Log("Cluster created")

	info, err = driver.PostCheck(ctx, info)
	assert.NoError(t, err)
	t.Log("Post Check done")

	newDriverOptions := getDriverOpts(t)
	newDriverOptions.IntOptions = GetNewIntOpts()

	t.Log("Update cluster by adding 3 nodes")
	newCount := int64(defaultNodeCount + 3)
	newDriverOptions.IntOptions["nodeCount"] = newCount
	info, err = driver.Update(ctx, info, newDriverOptions)
	assert.NoError(t, err)
	assert.EqualValues(t, newCount, info.NodeCount)
	realState := getRealClusterState(t, info)
	assert.EqualValues(t, newCount, len(realState.NodeIDs))
	t.Log("Resize is done: +3")

	t.Log("Update cluster by decreasing 2 nodes")
	newCount -= 2
	newDriverOptions.IntOptions["nodeCount"] = newCount
	info, err = driver.Update(ctx, info, newDriverOptions)
	assert.NoError(t, err)
	assert.EqualValues(t, newCount, info.NodeCount)
	realState = getRealClusterState(t, info)
	assert.EqualValues(t, newCount, len(realState.NodeIDs))
	t.Log("Resize is done: -2")

	t.Log("Start cluster removal")
	err = driver.Remove(ctx, info)
	assert.NoError(t, err)
}
