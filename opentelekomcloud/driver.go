package opentelekomcloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opentelekomcloud-infra/crutch-house/services"
	"github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/clusters"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodes"
	"github.com/rancher/kontainer-engine/drivers/util"
	"github.com/rancher/kontainer-engine/types"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	retries        = 5
	pollInterval   = 30
	baseServiceURL = "otc.t-systems.com"
)

var (
	authURL         = fmt.Sprintf("https://iam.eu-de.%s/v3", baseServiceURL)
	clusterVersions = []string{
		"v1.15.6-r1",
		"v1.13.10-r0",
	}
	clusterFlavors = []string{
		"cce.s1.small",
		"cce.s1.medium",
		"cce.s1.large",
		"cce.s2.small",
		"cce.s2.medium",
		"cce.s2.large",
		"cce.t1.small",
		"cce.t1.medium",
		"cce.t1.large",
		"cce.t2.small",
		"cce.t2.medium",
		"cce.t2.large",
	}
)

type managedResources struct {
	Vpc        bool
	Subnet     bool
	Cluster    bool
	Nodes      bool
	ClusterEip bool
}

type clusterState struct {
	types.ClusterInfo
	ClusterID             string
	AuthInfo              openstack.AuthInfo
	ClusterName           string
	DisplayName           string
	Description           string
	ProjectName           string
	Region                string
	ClusterType           string
	ClusterFlavor         string
	ClusterBillingMode    int
	ClusterLabels         map[string]string
	ContainerNetworkMode  string
	ContainerNetworkCidr  string
	VpcID                 string
	VpcName               string
	SubnetID              string
	SubnetName            string
	HighwaySubnetID       string
	HighwaySubnetName     string
	AuthenticatingProxyCa string
	UseFloatingIP         bool
	ClusterFloatingIP     string
	ClusterEIPOptions     services.ElasticIPOpts
	ClusterJobID          string
	NodeConfig            services.CreateNodesOpts
	NodeIDs               []string
	AuthMode              string
	ManagedResources      managedResources
}

type CCEDriver struct {
	driverCapabilities types.Capabilities
}

func (d *CCEDriver) GetDriverCreateOptions(context.Context) (*types.DriverFlags, error) {
	logrus.Info("Getting driver create opts...")
	flags := &types.DriverFlags{
		Options: map[string]*types.Flag{
			// Cluster general options
			"name": {
				Type:  types.StringType,
				Usage: "Cluster name",
			},
			"display-name": {
				Type:  types.StringType,
				Usage: "Cluster name displayed to user",
			},
			"description": {
				Type:  types.StringType,
				Usage: "Cluster description",
			},
			// Authentication options
			"domain-name": {
				Type:  types.StringType,
				Usage: "OTC domain name",
			},
			"project-name": {
				Type:  types.StringType,
				Usage: "OTC project name",
			},
			"username": {
				Type:  types.StringType,
				Usage: "OTC username",
			},
			"password": {
				Type:     types.StringType,
				Usage:    "OTC user password",
				Password: true,
			},
			"access-key": {
				Type:     types.StringType,
				Usage:    "OTC access key ID",
				Password: true,
			},
			"secret-key": {
				Type:     types.StringType,
				Usage:    "OTC secret access key",
				Password: true,
			},
			"token": {
				Type:     types.StringType,
				Usage:    "OTC token",
				Password: true,
			},
			"region": {
				Type:  types.StringType,
				Usage: "OTC region",
				Default: &types.Default{
					DefaultString: "eu-de",
				},
			},
			// Cluster configuration
			// cluster general settings
			"cluster-type": {
				Type:  types.StringType,
				Usage: "Type of the cluster, 'VirtualMachine' or 'BareMetal'",
				Default: &types.Default{
					DefaultString: "VirtualMachine",
				},
			},
			"cluster-version": {
				Type:  types.StringType,
				Usage: fmt.Sprintf("Version of k8s (one of %s), default is latest available", strings.Join(clusterVersions, ", ")),
			},
			"cluster-flavor": {
				Type:  types.StringType,
				Usage: "Cluster flavor, one of " + strings.Join(clusterFlavors, ", "),
				Default: &types.Default{
					DefaultString: "cce.s2.small",
				},
			},
			"cluster-billing-mode": {
				Type:  types.IntType,
				Usage: "The bill mode of the cluster",
				Default: &types.Default{
					DefaultInt: 0,
				},
			},
			"cluster-labels": {
				Type:  types.StringSliceType,
				Usage: "The map of Kubernetes labels (key/value pairs) to be applied to cluster",
			},
			// cluster networking
			"vpc": {
				Type:  types.StringType,
				Usage: "The name of VPC",
			},
			"vpc-id": {
				Type:  types.StringType,
				Usage: "The ID of existing VPC",
			},
			"subnet": {
				Type:  types.StringType,
				Usage: "The name of subnet",
			},
			"subnet-id": {
				Type:  types.StringType,
				Usage: "The ID of existing subnet",
			},
			"highway-subnet": {
				Type:  types.StringType,
				Usage: "The id of existing highway subnet when the cluster-type is BareMetal",
			},
			"container-network-mode": {
				Type:  types.StringType,
				Usage: "The network mode of container",
				Value: "overlay_l2",
			},
			"container-network-cidr": {
				Type:  types.StringType,
				Usage: "The network cidr of container",
				Default: &types.Default{
					DefaultString: "172.16.0.0/16",
				},
			},
			// cluster auth
			"authentication-mode": {
				Type:  types.StringType,
				Usage: "The Authentication Mode for cce cluster. rbac or authenticating_proxy, default to rbac",
				Default: &types.Default{
					DefaultString: "rbac",
				},
			},
			"auth-proxy-ca": {
				Type:  types.StringType,
				Usage: "The CA for authenticating proxy, it is required if authentication-mode is authenticating_proxy",
			},
			"cluster-floating-ip": {
				Type:  types.StringType,
				Usage: "Existing floating IP to be associated with cluster master node",
			},
			// Nodes configuration
			"node-count": {
				Type:  types.IntType,
				Usage: "The number of nodes to create in this cluster",
			},
			"availability-zone": {
				Type:  types.StringType,
				Usage: "AZ used for node creation",
				Default: &types.Default{
					DefaultString: "eu-de-01",
				},
			},
			"node-flavor": {
				Type:  types.StringType,
				Usage: "The node flavor",
				Default: &types.Default{
					DefaultString: "s3.large.2",
				},
			},
			"node-os": {
				Type:  types.StringType,
				Usage: "The operation system of nodes",
				Default: &types.Default{
					DefaultString: "EulerOS 2.5",
				},
			},
			"key-pair": {
				Type:  types.StringType,
				Usage: "The name of ssh key-pair",
			},
			// BMS settings
			"billing-mode": {
				Type:    types.IntType,
				Usage:   "The bill mode of the nodes",
				Default: &types.Default{DefaultInt: 0},
			},
			"bms-period-type": {
				Type:    types.StringType,
				Usage:   "The period type",
				Default: &types.Default{DefaultString: "month"},
			},
			"bms-period-num": {
				Type:    types.IntType,
				Usage:   "The number of period",
				Default: &types.Default{DefaultInt: 1},
			},
			"bms-auto-renew": {
				Type:  types.BoolType,
				Usage: "If the period is auto renew",
			},
			// disk settings
			"root-volume-size": {
				Type:    types.IntType,
				Usage:   "Size of the system disk attached to each node in GB, 40 min",
				Default: &types.Default{DefaultInt: 40},
			},
			"root-volume-type": {
				Type:    types.StringType,
				Usage:   "Type of the system disk attached to each node, one of SATA, SAS, SSD",
				Default: &types.Default{DefaultString: "SATA"},
			},
			"data-volume-size": {
				Type:    types.IntType,
				Usage:   "Size of the data disk attached to each node in GB, 100 min",
				Default: &types.Default{DefaultInt: 100},
			},
			"data-volume-type": {
				Type:    types.StringType,
				Usage:   "Type of the data disk attached to each node, one of SATA, SAS, SSD",
				Default: &types.Default{DefaultString: "SATA"},
			},
			// master node bandwidth
			"cluster-eip-type": {
				Type:    types.StringType,
				Usage:   "The type of bandwidth",
				Default: &types.Default{DefaultString: "5_bgp"},
			},
			"cluster-eip-bandwidth-size": {
				Type:    types.IntType,
				Usage:   "The size of bandwidth, MBit",
				Default: &types.Default{DefaultInt: 100},
			},
			"cluster-eip-share-type": {
				Type:    types.StringType,
				Usage:   "The share type of bandwidth",
				Default: &types.Default{DefaultString: "PER"},
			},
			// lb
			"load-balancer": {
				Type:    types.StringType,
				Usage:   "Existing LB ID",
				Default: &types.Default{DefaultString: ""},
			},
		},
	}
	return flags, nil
}

func (d *CCEDriver) GetDriverUpdateOptions(context.Context) (*types.DriverFlags, error) {
	flags := &types.DriverFlags{
		Options: map[string]*types.Flag{
			// Cluster general options
			"description": {
				Type:  types.StringType,
				Usage: "Cluster description",
			},
		},
	}
	return flags, nil
}

func stateFromOpts(opts *types.DriverOptions) (*clusterState, error) {
	logrus.Info("Start setting state from provided opts: \n", opts)
	strOpt, strSliceOpt, intOpt, boolOpt := getters(opts)
	projectName := strOpt("project-name", "projectName")
	state := &clusterState{
		ClusterInfo: types.ClusterInfo{
			Version:   strOpt("cluster-version", "clusterVersion"),
			NodeCount: intOpt("node-count", "nodeCount"),
		},
		AuthInfo: openstack.AuthInfo{
			AuthURL:     authURL,
			Token:       strOpt("token"),
			Username:    strOpt("username"),
			Password:    strOpt("password"),
			ProjectName: projectName,
			DomainName:  strOpt("domain-name", "domainName"),
			AccessKey:   strOpt("access-key", "accessKey"),
			SecretKey:   strOpt("secret-key", "secretKey"),
		},
		ClusterName:           strOpt("name"),
		DisplayName:           strOpt("display-name", "displayName"),
		Description:           strOpt("description"),
		ProjectName:           strOpt("project-name", "projectName"),
		Region:                strOpt("region"),
		ClusterType:           strOpt("cluster-type", "clusterType"),
		ClusterFlavor:         strOpt("cluster-flavor", "clusterFlavor"),
		ClusterBillingMode:    int(intOpt("cluster-billing-mode", "clusterBillingMode")),
		ClusterLabels:         map[string]string{},
		ContainerNetworkMode:  strOpt("container-network-mode", "containerNetworkMode"),
		ContainerNetworkCidr:  strOpt("container-network-cidr", "containerNetworkCidr"),
		AuthenticatingProxyCa: strOpt("auth-proxy-ca", "authProxyCa"),
		UseFloatingIP:         !boolOpt("no-floating-ip", "noFloatingIp"),
		ClusterFloatingIP:     strOpt("cluster-floating-ip", "clusterFloatingIp"),
		ClusterEIPOptions: services.ElasticIPOpts{
			IPType:        strOpt("cluster-eip-type", "clusterEipType"),
			BandwidthSize: int(intOpt("cluster-eip-bandwidth-size", "clusterEipBandwidthSize")),
			BandwidthType: strOpt("cluster-eip-share-type", "clusterEipShareType"),
		},

		NodeConfig: services.CreateNodesOpts{
			Region:           projectName,
			FlavorID:         strOpt("node-flavor", "nodeFlavor"),
			AvailabilityZone: strOpt("availability-zone", "availabilityZone"),
			KeyPair:          strOpt("key-pair", "keyPair"),
			RootVolume: nodes.VolumeSpec{
				Size:       int(intOpt("root-volume-size", "rootVolumeSize")),
				VolumeType: strOpt("root-volume-type", "rootVolumeType"),
			},
			DataVolumes: []nodes.VolumeSpec{
				{
					Size:       int(intOpt("data-volume-size", "dataVolumeSize")),
					VolumeType: strOpt("data-volume-type", "dataVolumeType"),
				},
			},
			Os:       strOpt("node-os", "os", "nodeOs"),
			EipCount: 0,
		},
		AuthMode:          strOpt("auth-mode", "authenticationMode"),
		VpcName:           strOpt("vpc", "vpcName"),
		VpcID:             strOpt("vpc-id", "vpcId"),
		SubnetName:        strOpt("subnet", "subnetName"),
		SubnetID:          strOpt("subnet-id", "subnetId"),
		HighwaySubnetName: strOpt("highway-subnet", "highwaySubnetName"),
	}

	for _, label := range strSliceOpt("cluster-labels", "clusterLabels") {
		lab := strings.Split(label, "=")
		if len(lab) != 2 {
			return nil, fmt.Errorf("invalid label value: %s", label)
		}
		state.ClusterLabels[lab[0]] = lab[1]
	}

	return state, nil
}

// Load state from in ClusterInfo Metadata
func stateFromInfo(info *types.ClusterInfo) (*clusterState, error) {
	state := &clusterState{}

	err := json.Unmarshal([]byte(info.Metadata["state"]), state)
	if err != nil {
		logrus.Errorf("Error encountered while marshalling state: %v", err)
	}

	return state, err
}

// Save state to ClusterInfo Metadata
func stateToInfo(info *types.ClusterInfo, state clusterState) error {
	data, err := json.Marshal(state)

	if err != nil {
		return err
	}

	if info.Metadata == nil {
		info.Metadata = map[string]string{}
	}

	info.Metadata["state"] = string(data)

	return nil
}

func setupNetwork(client services.Client, state *clusterState) error {
	logrus.Debug("Setup network process started")
	if state.VpcID == "" && state.VpcName != "" {
		vpcID, err := client.FindVPC(state.VpcName)
		if err != nil {
			return err
		}
		if vpcID == "" {
			vpc, err := client.CreateVPC(state.VpcName)
			if err != nil {
				return err
			}
			state.ManagedResources.Vpc = true
			vpcID = vpc.ID
		}
		if err := client.WaitForVPCStatus(vpcID, "OK"); err != nil {
			return fmt.Errorf("failed waiting for VPC status 'OK': %s", err)
		}
		state.VpcID = vpcID
	}

	if state.SubnetID == "" && state.SubnetName != "" {
		subnetID, err := client.FindSubnet(state.VpcID, state.SubnetName)
		if err != nil {
			return err
		}
		if subnetID == "" {
			subnet, err := client.CreateSubnet(state.VpcID, state.SubnetName)
			if err != nil {
				return err
			}
			state.ManagedResources.Subnet = true
			subnetID = subnet.ID
		}
		if err := client.WaitForSubnetStatus(subnetID, "ACTIVE"); err != nil {
			return fmt.Errorf("failed wating for subnet sttatus 'ACTIVE': %s", err)
		}
		state.SubnetID = subnetID
	}

	if state.HighwaySubnetID == "" && state.HighwaySubnetName != "" {
		highwaySubnetID, err := client.FindSubnet(state.VpcID, state.HighwaySubnetName)
		if err != nil {
			return err
		}
		state.HighwaySubnetName = highwaySubnetID
	}

	if state.ClusterFloatingIP == "" {
		eip, err := client.CreateEIP(&state.ClusterEIPOptions)
		if err != nil {
			return err
		}
		state.ManagedResources.ClusterEip = true
		state.ClusterFloatingIP = eip.PublicAddress
	}

	logrus.Debug("Setup network process finished")
	return nil
}

func createCluster(client services.Client, state *clusterState) error {
	var nodeIPs []string
	var nodeIDs []string
	var clusterID string
	cluster, err := client.CreateCluster(&services.CreateClusterOpts{
		Name:            state.ClusterName,
		Description:     state.Description,
		ClusterType:     state.ClusterType,
		ClusterVersion:  state.Version,
		FlavorID:        state.ClusterFlavor,
		VpcID:           state.VpcID,
		SubnetID:        state.SubnetID,
		HighwaySubnetID: state.HighwaySubnetID,
		ContainerNetwork: clusters.ContainerNetworkSpec{
			Mode: state.ContainerNetworkMode,
			Cidr: state.ContainerNetworkCidr,
		},
		AuthenticationMode: state.AuthMode,
		BillingMode:        state.ClusterBillingMode,
		FloatingIP:         state.ClusterFloatingIP,
	})
	if err != nil {
		return err
	}
	clusterID = cluster.Metadata.Id
	state.ClusterID = clusterID
	state.NodeConfig.ClusterID = clusterID

	nodeIDs, err = client.CreateNodes(&state.NodeConfig, int(state.NodeCount))
	if err != nil {
		return fmt.Errorf("failed to create nodes: %s", err)
	}
	state.NodeIDs = nodeIDs
	nodesStatus, err := client.GetNodesStatus(clusterID, nodeIDs)
	if err != nil {
		return err
	}
	nodeIPs = make([]string, len(nodeIDs))
	for i, nodeStatus := range nodesStatus {
		nodeIPs[i] = nodeStatus.PrivateIP
	}
	return nil
}

func getClient(state *clusterState) (client services.Client, err error) {
	client = services.NewCloudClient(&openstack.Cloud{
		AuthInfo:     state.AuthInfo,
		RegionName:   state.Region,
		EndpointType: "public",
	})
	if err := client.Authenticate(); err != nil {
		return nil, err
	}
	if err := client.InitVPC(); err != nil {
		return nil, err
	}
	if err := client.InitNetworkV2(); err != nil {
		return nil, err
	}
	if err := client.InitCompute(); err != nil {
		return nil, err
	}
	if err := client.InitCCE(); err != nil {
		return nil, err
	}
	return client, nil
}

func cleanupManagedResources(client services.Client, state *clusterState) error {
	logrus.Debug("Cleanup process started")
	resources := state.ManagedResources

	if resources.Nodes {
		resources.Nodes = false
	}
	if resources.Cluster {
		resources.Cluster = false
	}
	if resources.ClusterEip {
		if err := client.DeleteFloatingIP(state.ClusterFloatingIP); err != nil {
			return err
		}
		resources.ClusterEip = false
	}
	if resources.Subnet {
		if err := client.DeleteSubnet(state.VpcID, state.SubnetID); err != nil {
			return err
		}
		err := client.WaitForSubnetStatus(state.SubnetID, "")
		if err, ok := err.(golangsdk.ErrDefault404); !ok {
			return err
		}
		resources.Subnet = false
	}
	if resources.Vpc {
		if err := client.DeleteVPC(state.VpcID); err != nil {
			return err
		}
		resources.Vpc = false
	}
	logrus.Debug("Cleanup process finished")
	return nil
}

func (d *CCEDriver) Create(_ context.Context, opts *types.DriverOptions, _ *types.ClusterInfo) (clusterInfo *types.ClusterInfo, err error) {
	logrus.Info("Start creating cluster")
	logrus.Debug("Get state from opts")
	state, err := stateFromOpts(opts)
	if err != nil {
		return nil, err
	}

	info := &types.ClusterInfo{}
	defer func() {
		logrus.WithError(storeState(info, state)).Info("Save cluster state: ", state)
	}()
	client, err := getClient(state)
	if err != nil {
		return nil, err
	}
	token, _ := client.Token() // error can only be because of auth
	state.ServiceAccountToken = token

	state.ManagedResources = managedResources{}
	defer func() {
		if err != nil {
			logrus.WithError(err).Info("creation failed")
			logrus.WithError(cleanupManagedResources(client, state))
		}
	}()

	if err := setupNetwork(client, state); err != nil {
		return nil, fmt.Errorf("failed to setup network: %s", err)
	}

	if err := createCluster(client, state); err != nil {
		return nil, fmt.Errorf("failed to create cluster: %s", err)
	}

	logrus.Info("Cluster creation finished")
	return info, stateToInfo(info, *state)
}

// Update changes existing cluster. `clusterInfo` represents current state, `updateOpts` are newly applied flags
func (d *CCEDriver) Update(_ context.Context, info *types.ClusterInfo, updateOpts *types.DriverOptions) (*types.ClusterInfo, error) {
	var err error
	defer func() {
		if err != nil {
			logrus.WithError(err).Info("update return error")
		}
	}()
	logrus.Info("Starting update")
	state, err := stateFromInfo(info)
	if err != nil {
		return nil, err
	}

	newState, err := stateFromOpts(updateOpts)
	if err != nil {
		return nil, err
	}
	newState.ClusterID = state.ClusterID

	if newState.NodeCount != state.NodeCount {
		tmpState, err := d.resizeCluster(info, newState.NodeCount)
		if err != nil {
			return nil, err
		}
		state.NodeIDs = tmpState.NodeIDs
		state.NodeCount = newState.NodeCount
	}

	if newState.Description != state.Description {
		client, err := getClient(state)
		if err != nil {
			return nil, err
		}
		spec := &clusters.UpdateSpec{Description: newState.Description}
		if err := client.UpdateCluster(newState.ClusterID, spec); err != nil {
			return nil, err
		}
		state.Description = newState.Description
	}

	logrus.Info("update cluster success")
	return info, stateToInfo(info, *state)
}

func (d *CCEDriver) PostCheck(_ context.Context, clusterInfo *types.ClusterInfo) (*types.ClusterInfo, error) {
	state, err := stateFromInfo(clusterInfo)
	if err != nil {
		return nil, err
	}
	client, err := getClient(state)
	if err != nil {
		return nil, err
	}
	cluster, err := client.GetCluster(state.ClusterID)
	if err != nil {
		return nil, err
	}
	if logrus.GetLevel() == logrus.DebugLevel {
		jsonData, _ := json.Marshal(cluster)
		logrus.Debugf("cluster info %s", string(jsonData))
	}

	cert, err := client.GetClusterCertificate(state.ClusterID)
	if err != nil {
		return nil, err
	}
	if logrus.GetLevel() == logrus.DebugLevel {
		jsonData, _ := json.Marshal(cert)
		logrus.Debugf("cert info %s", string(jsonData))
	}

	for _, cluster := range cert.Clusters {
		switch cluster.Name {
		case "internalCluster":
			clusterInfo.RootCaCertificate = cluster.Cluster.CertAuthorityData
			break
		case "externalCluster":
			clusterInfo.Endpoint = cluster.Cluster.Server
			break
		}
	}

	state.Endpoint = clusterInfo.Endpoint

	clusterInfo.Status = cluster.Status.Phase
	clusterInfo.ClientKey = cert.Users[0].User.ClientKeyData
	clusterInfo.ClientCertificate = cert.Users[0].User.ClientCertData
	clusterInfo.Username = cert.Users[0].Name

	clientSet, err := getClientSet(clusterInfo)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %v", err)
	}

	failureCount := 0

	for {
		clusterInfo.ServiceAccountToken, err = util.GenerateServiceAccountToken(clientSet)

		if err == nil {
			logrus.Info("service account token generated successfully")
			break
		} else {
			logrus.WithError(err).Warnf("error creating service account")
			if failureCount < retries {
				logrus.Infof("service account token generation failed, retries left: %v", retries-failureCount)
				failureCount = failureCount + 1

				time.Sleep(pollInterval * time.Second)
			} else {
				logrus.Error("retries exceeded, failing post-check")
				return nil, err
			}
		}
	}
	logrus.Info("post-check completed successfully")
	logrus.Debugf("info: %v", *clusterInfo)

	return clusterInfo, nil
}

func (d *CCEDriver) Remove(_ context.Context, clusterInfo *types.ClusterInfo) error {
	logrus.Info("Get state from info")
	state, err := stateFromInfo(clusterInfo)
	if err != nil {
		return err
	}
	client, err := getClient(state)
	if err != nil {
		return err
	}
	if err := client.DeleteNodes(state.ClusterID, state.NodeIDs); err != nil {
		return err
	}
	if err := client.DeleteCluster(state.ClusterID); err != nil {
		return err
	}
	if err := cleanupManagedResources(client, state); err != nil {
		return err
	}
	return nil
}

func (d *CCEDriver) GetVersion(_ context.Context, info *types.ClusterInfo) (*types.KubernetesVersion, error) {
	state, err := stateFromInfo(info)
	if err != nil {
		return nil, err
	}
	return &types.KubernetesVersion{Version: state.Version}, nil
}

func (d *CCEDriver) SetVersion(context.Context, *types.ClusterInfo, *types.KubernetesVersion) error {
	return fmt.Errorf("setting version is not implemented")
}

func (d *CCEDriver) GetClusterSize(_ context.Context, info *types.ClusterInfo) (*types.NodeCount, error) {
	state, err := stateFromInfo(info)
	if err != nil {
		return nil, err
	}
	return &types.NodeCount{Count: state.NodeCount}, nil
}

func (d *CCEDriver) resizeCluster(info *types.ClusterInfo, newSize int64) (*clusterState, error) {
	state, err := stateFromInfo(info)
	if err != nil {
		return nil, err
	}
	client, err := getClient(state)
	if err != nil {
		return nil, err
	}
	delta := newSize - state.NodeCount
	logrus.Info("Start setting cluster size")
	if delta == 0 {
		return state, nil
	}
	if delta > 0 {
		state.NodeConfig.ClusterID = state.ClusterID
		newNodes, err := client.CreateNodes(&state.NodeConfig, int(delta))
		if err != nil {
			return nil, err
		}
		state.NodeIDs = append(state.NodeIDs, newNodes...)
	} else {
		nodesToDelete := state.NodeIDs[newSize:]
		err := client.DeleteNodes(state.ClusterID, nodesToDelete)
		if err != nil {
			return nil, err
		}
		state.NodeIDs = state.NodeIDs[:newSize]
	}
	if len(state.NodeIDs) != int(newSize) {
		return nil, fmt.Errorf("resize failed: expected %v items in %v", newSize, state.NodeIDs)
	}
	state.NodeCount = newSize
	logrus.Infof("Setting cluster size to %v finished", newSize)
	return state, nil
}

func (d *CCEDriver) SetClusterSize(_ context.Context, info *types.ClusterInfo, count *types.NodeCount) error {
	_, err := d.resizeCluster(info, count.Count)
	if err != nil {
		return err
	}
	return nil
}

func (d *CCEDriver) GetCapabilities(context.Context) (*types.Capabilities, error) {
	return &d.driverCapabilities, nil
}

func getClientSet(info *types.ClusterInfo) (clientSet *kubernetes.Clientset, err error) {
	certBytes, err := base64.StdEncoding.DecodeString(info.ClientCertificate)
	if err != nil {
		return nil, err
	}
	keyBytes, err := base64.StdEncoding.DecodeString(info.ClientKey)
	if err != nil {
		return nil, err
	}
	rootBytes, err := base64.StdEncoding.DecodeString(info.RootCaCertificate)
	if err != nil {
		return nil, err
	}
	config := &rest.Config{
		Host: info.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   rootBytes,
			CertData: certBytes,
			KeyData:  keyBytes,
		},
	}
	return kubernetes.NewForConfig(config)
}

func (d *CCEDriver) RemoveLegacyServiceAccount(_ context.Context, info *types.ClusterInfo) error {
	clientSet, err := getClientSet(info)
	if err != nil {
		return err
	}
	return util.DeleteLegacyServiceAccountAndRoleBinding(clientSet)
}

var noETCDBackup = fmt.Errorf("ETCD backup operations are not implemented")

func (d *CCEDriver) ETCDSave(context.Context, *types.ClusterInfo, *types.DriverOptions, string) error {
	return noETCDBackup
}

func (d *CCEDriver) ETCDRestore(context.Context, *types.ClusterInfo, *types.DriverOptions, string) (*types.ClusterInfo, error) {
	return nil, noETCDBackup
}

func (d *CCEDriver) ETCDRemoveSnapshot(context.Context, *types.ClusterInfo, *types.DriverOptions, string) error {
	return noETCDBackup
}

func (d *CCEDriver) GetK8SCapabilities(context.Context, *types.DriverOptions) (*types.K8SCapabilities, error) {
	return &types.K8SCapabilities{
		L4LoadBalancer: &types.LoadBalancerCapabilities{
			Enabled: false,
		},
		NodePoolScalingSupported: false,
	}, nil
}

func NewDriver() types.Driver {
	driver := &CCEDriver{
		driverCapabilities: types.Capabilities{
			Capabilities: make(map[int64]bool),
		},
	}

	driver.driverCapabilities.AddCapability(types.GetVersionCapability)
	driver.driverCapabilities.AddCapability(types.GetClusterSizeCapability)
	driver.driverCapabilities.AddCapability(types.SetClusterSizeCapability)

	return driver
}

func storeState(info *types.ClusterInfo, state *clusterState) error {
	bytes, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if info.Metadata == nil {
		info.Metadata = map[string]string{}
	}
	info.Metadata["state"] = string(bytes)
	return nil
}
