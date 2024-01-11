Kontainer Engine Driver OTC
===========================
[![Go Report Card](https://goreportcard.com/badge/github.com/opentelekomcloud/kontainer-engine-driver-otc)](https://goreportcard.com/report/github.com/opentelekomcloud/kontainer-engine-driver-otc)
![GitHub](https://img.shields.io/github/license/opentelekomcloud/kontainer-engine-driver-otc)
[![Build Status](https://zuul.otc-service.com/api/tenant/eco/badge?project=opentelekomcloud/kontainer-engine-driver-otc&pipeline=gate)](https://zuul.otc-service.com/t/eco/builds?project=opentelekomcloud%2Fkontainer-engine-driver-otc&pipeline=gate)
![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/opentelekomcloud/kontainer-engine-driver-otc?include_prereleases)

This repository contains the Open Telekom Cloud CCE cluster driver (a. k. a. the "Kontainer-Engine-Driver") for the [Rancher](https://www.rancher.com/) Enterprise Kubernetes Management platform.

Frontend part: 
 - https://github.com/opentelekomcloud/ui-cluster-driver-otc

## Prerequisites

This driver engine for the Open Telekom Cloud assumes that you have already the **cluster server** of the Rancher Enterprise Kubernetes management platform installed, either on one or several bare-metal-servers, in a Kubernetes cluster, or even in a single container. All options and a thorough discussion of all alternatives and options are covered by the [server installation manual](https://ranchermanager.docs.rancher.com/pages-for-subheaders/installation-and-upgrade).

This driver is tested with **Rancher 2.7.9**. Any newer version should do. You need to create at least one user on the Rancher cluster server, a password, and the Rancher server log-in URL. For details, refer to the Rancher manual.

If you want to utilize **infrastructure resources provisioned by the Open Telekom Cloud**, you need exactly this very driver. The compute, network, and storage resources covered by the Open Telekom Cloud can be consumed and managed by the Rancher platform. You thus need **access credentials** for the Open Telekom Cloud. They come usually in form of a `clouds.yaml` file or an AK/SK pair.

Depending on your cluster setup, you may need **preconfigured network resources** such as a VPC, with at least one network, Internet connectivity for example via a NAT gateway or a router configuration associated with an EIP. You should upload an **SSH key pair** before you configure a new cluster. Please refer for all of this to your [Open Telekom Cloud documentation](https://docs.otc.t-systems.com/). These few tasks can be performed either via the [Open Telekom Cloud browser-based console](https://console.otc.t-systems.com/) or with a CLI tool like the [OpenStack Client](https://docs.openstack.org/python-openstackclient/).
  
## Building

To build the driver from its Go sources, run `make build` from the main directory of this repository. 
The output driver binaries are dropped into the `bin` directory. These files can be imported 
directly into Rancher and used as cluster drivers.  They must be distributed 
via URLs that your Rancher instance can establish a connection to and download 
the driver binaries. For your convenience, this driver is distributed via a GitHub 
release and can be downloaded from one of those download URLs directly.

## Adding or Updating a Driver

1. Open your `Rancher` UI website and log into its service.

2. Go to the `Cluster Drivers` management screen in Rancher.

   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/cluster-mgmt.png" alt="image" style="width:150px;height:auto;">

3. Go to `Drivers` -> `Cluster Drivers` and search for `Open Telekom Cloud CCE`. Click the three dotted menu and then `Edit`.

   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/cluster_drivers.png" alt="image" style="width:1000px;height:auto;">

4. Alternatively just click `Add Cluster Driver`.

5. Enter Download URL:
    1) Using exact version: `https://otc-rancher.obs.eu-de.otc.t-systems.com/cluster/driver/1.1.1/kontainer-engine-driver-otccce_linux_amd64.tar.gz`
    2) Using the latest version: `https://otc-rancher.obs.eu-de.otc.t-systems.com/cluster/driver/latest/kontainer-engine-driver-otccce_linux_amd64.tar.gz`

6. Enter the Custom UI URL:
   1) Using exact version: `https://otc-rancher.obs.eu-de.otc.t-systems.com/cluster/ui/v1.2.0/component.js`.
   2) Or latest: `https://otc-rancher.obs.eu-de.otc.t-systems.com/cluster/ui/latest/component.js`.

7. Add Whitelist Domains with value `*.otc.t-systems.com`.

   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/edit_cluster_driver.png" alt="image" style="width:600px;height:auto;">

8. Click `Save` if you are in edit mode of existing driver or `Create` for new one, and wait for driver status to be `Active`.

9. Cluster driver for OpenTelekomCloud CCE service will be available to use on the `Cluster:Create` screen.

   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/cluster_create.png" alt="image" style="width:1000px;height:auto;">

## Creating a Cluster

> To establish a successful connection from yor Rancher agents to your cluster, you have to provide Internet access to your nodes. You can do that for example by creating a `NAT Gateway` for the VPC that hosts your cluster with either the CLI or the Console of the Open Telekom Cloud. Then add an `SNAT` rule for your `VPC` and `subnet` and pass the `EIP`.

> **WAITING FOR FIX** from **RANCHER**: Currently after cluster provision you must change `fleet-agent` pod image from default `image: rancher/fleet-agent:v0.8.1` to `image: rancher/fleet-agent:v0.9.0`

1. Go to `Clusters` and click `Create`.
   
2. Click on `Open Telekom Cloud CCE`.

3. Choose authentication method `AK/SK` or `Token-based`. The latter is recommended uses the same values and formats as configured in a `clouds.yaml` file. Fill in the required fields, such as `Region`, `Domain Name`, `Project Name`, `Username`, `Password`, `Access Key Id`, `Secret Access Key`.
   
   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/auth.png" alt="image" style="width:800px;height:auto;">

   Then click `Next: Configure Cluster`.

4. On `Cluster Configuration` choose `Kubernetes version`
   > **Supported versions are 1.23 and 1.25**)

   `Cluster flavor`, `Network mode` and `CIDR`.

   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/cluster.png" alt="image" style="width:800px;height:auto;">
   
   Then click `Next: Network Configuration`.
   
5. On `Network Configuration` choose `VPC` from list of created Vpcs and choose `Subnet`.
   
   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/network.png" alt="image" style="width:800px;height:auto;">
   
   Then click `Next: Cluster Floating IP`.
   
6. On `Cluster Floating IP` you can create new IP with selected bandwidth size or use existing one.
   
   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/ip.png" alt="image" style="width:800px;height:auto;">
   
   Then click `Next: Node Configuration`.
   
7. On `Node Configuration` choose `Node Count` 
   > Select at least three or more nodes, since Rancher wants a lot of resources.

   `Node OS`, `Availability Zone`, `Key Pair` and `Node Flavor`
   > `Node Flavor` list depends on a selected `Availability Zone`, so please choose that first.

   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/node.png" alt="image" style="width:800px;height:auto;">
   
   Then click `Next: Nodes disk configuration`.
   
8. On `Node Configuration` select `Root Disk Type` and `Data Disk Type` and `sizes`.

   <img src="https://otc-rancher.obs.eu-de.otc.t-systems.com/helpers/disk.png" alt="image" style="width:800px;height:auto;">

   Then click `Finish & Create Cluster`. 

> Sometimes `Rancher` doesn't display the new cluster until all cluster node have been provisioned. If that happens, please wait and check the console. (WHERE/HOW?)

## License

Copyright © 2023, 2024 by Open Telekom Cloud, T-Systems International GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
