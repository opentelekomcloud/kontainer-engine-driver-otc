Kontainer Engine Driver OTC
===============================
[![Go Report Card](https://goreportcard.com/badge/github.com/opentelekomcloud/kontainer-engine-driver-otc)](https://goreportcard.com/report/github.com/opentelekomcloud/kontainer-engine-driver-otc)
![GitHub](https://img.shields.io/github/license/opentelekomcloud/kontainer-engine-driver-otc)

This repo contains the OTC CCE(Open Telekom Cloud Container Engine) driver for the rancher server.

## Building

`make build`

Will output driver binaries into the `bin` directory, these can be imported 
directly into Rancher and used as cluster drivers.  They must be distributed 
via URLs that your Rancher instance can establish a connection to and download 
the driver binaries.  For example, this driver is distributed via a GitHub 
release and can be downloaded from one of those URLs directly.


## Running

1. Go to the `Cluster Drivers` management screen in Rancher and click `Add Cluster Driver`.
2. Enter Download URL `https://github.com/opentelekomcloud/kontainer-engine-driver-otc/releases/download/VERSION/kontainer-engine-driver-otc-VERSION-linux-amd64.tgz`
3. Enter the Custom UI URL with value `IN DEVELOPMENT PHASE`.
4. Add Whitelist Domains with value `*.otc.t-systems.com`.
5. Click `Create`, and wait for driver status to be `Active`.
6. OTC CCEDriver will be available to use on the `Add Cluster` screen.