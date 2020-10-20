Kontainer Engine Driver OTC
===============================
[![Go Report Card](https://goreportcard.com/badge/github.com/opentelekomcloud/kontainer-engine-driver-otc)](https://goreportcard.com/report/github.com/opentelekomcloud/kontainer-engine-driver-otc)
![GitHub](https://img.shields.io/github/license/opentelekomcloud/kontainer-engine-driver-otc)
[![Zuul Gate](https://zuul.eco.tsi-dev.otc-service.com/api/tenant/eco/badge?project=opentelekomcloud/kontainer-engine-driver-otc&pipeline=check&branch=master)](https://zuul.eco.tsi-dev.otc-service.com/t/eco/builds?project=opentelekomcloud%2Fkontainer-engine-driver-otc)
![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/opentelekomcloud/kontainer-engine-driver-otc?include_prereleases)

This repo contains the Open Telekom Cloud CCE driver for the rancher server.

## Building

`make build`

Will output driver binaries into the `bin` directory, these can be imported 
directly into Rancher and used as cluster drivers.  They must be distributed 
via URLs that your Rancher instance can establish a connection to and download 
the driver binaries.  For example, this driver is distributed via a GitHub 
release and can be downloaded from one of those URLs directly.


## Running

1. Go to the `Cluster Drivers` management screen in Rancher and click `Add Cluster Driver`.
2. Enter Download URL:
    1) Using exact version: `https://github.com/opentelekomcloud/kontainer-engine-driver-otc/releases/download/VERSION/kontainer-engine-driver-otc-VERSION-linux-amd64.tgz`
    2) Using latest version: `https://otc-rancher.obs.eu-de.otc.t-systems.com/cluster/driver/latest/kontainer-engine-driver-otccce_linux_amd64.tar.gz`
3. Enter the Custom UI URL: `https://otc-rancher.obs.eu-de.otc.t-systems.com/cluster/ui/latest/component.js`.
4. Add Whitelist Domains with value `*.otc.t-systems.com`.
5. Click `Create`, and wait for driver status to be `Active`.
6. Cluster driver for OpenTelekomCloud CCE service will be available to use on the `Add Cluster` screen.

## License

Copyright 2020 T-Systems GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
