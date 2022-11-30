[![Build Status](https://circleci.com/gh/treydock/infiniband_exporter/tree/master.svg?style=shield)](https://circleci.com/gh/treydock/infiniband_exporter)
[![GitHub release](https://img.shields.io/github/v/release/treydock/infiniband_exporter?include_prereleases&sort=semver)](https://github.com/treydock/infiniband_exporter/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/treydock/infiniband_exporter/total)
[![Go Report Card](https://goreportcard.com/badge/github.com/treydock/infiniband_exporter)](https://goreportcard.com/report/github.com/treydock/infiniband_exporter)
[![codecov](https://codecov.io/gh/treydock/infiniband_exporter/branch/master/graph/badge.svg)](https://codecov.io/gh/treydock/infiniband_exporter)

# InfiniBand Prometheus exporter

The InfiniBand exporter collects counters from InfiniBand switches and HCAs.
The exporter supports the `/metrics` endpoint to gather InfiniBand metrics and metrics about the exporter.

This exporter must be run on a host that has an active interface on the InfiniBand fabric you wish to monitor.
By default this exporter will collect counters from all switch ports on the fabric connected to the host running this exporter.

The InfiniBand diagnostic tools of `ibnetdiscover` and `perfquery` must also be present on the host running this exporter.
These are commonly installed via the `infiniband-diags` package.

## Usage

Collectors are enabled or disabled via `--collector.<name>` and `--no-collector.<name>` flags.

Name | Description | Default
-----|-------------|--------
switch | Collect switch port counters | Enabled
ibswinfo | Collect data on unmanaged switches via ibswinfo | Disabled
hca | Collect HCA port counters | Disabled

If you have a node name map file typically used with Subnet Managers, you can provide that file to the  `--ibnetdiscover.node-name-map` flag.  This will use friendly names for switches.


If you wish to run the exporter as a user other than root and do not want to use sudo, you must make the UMAD device read/write to all users with something like the following:

```
$ cat /etc/udev/rules.d/99-ib.rules 
KERNEL=="umad*", NAME="infiniband/%k" MODE="0666"
```

If you wish to use sudo you will need to run with the `--sudo` flag.  Below is an example of the sudo rules necessary if the exporter rules as `infiniband_exporter` user: (adjust paths to `perfquery` and `ibnetdiscover` as needed)

```
Defaults:infiniband_exporter !syslog
Defaults:infiniband_exporter !requiretty
infiniband_exporter ALL=(ALL) NOPASSWD: /usr/sbin/ibnetdiscover
infiniband_exporter ALL=(ALL) NOPASSWD: /usr/sbin/perfquery
```

If `ibnetdiscover` and `perfquery` are not in PATH then their paths need to be provided via the `--ibnetdiscover.path` and `--perfquery.path` flags.

### Collect switch information using ibswinfo (BETA)

The tool [ibswinfo](https://github.com/stanford-rc/ibswinfo) can be used to collect information from unmanaged InfiniBand switches such as power supply and fan health.  To enable this collection pass the `--collector.switch.ibswinfo` flag and ensure either `ibswinfo` is in $PATH or define the path to that executable via the `--ibswinfo.path` flag.

This feature is considered BETA as it relies on parsing non-machine readable data.
In the future this exporter may collect the unmanaged switch information directly in a similar way to what ibswinfo is doing.

### Large fabric considerations

If you have a large fabric where collection times are too long for Prometheus scrapes, the exporter can instead write metrics to a file that can be collected by node_exporter textfile collection.

This exporter has been tested on a fabric with 109 switches each having around 36 ports and collecting only switches takes ~10 seconds.

To collect the metrics from a file pass the `--collector.textfile.directory` flag to node_exporter like so: `--collector.textfile.directory=/var/lib/node_exporter/textfile_collector`.  Add this exporter to be executed via cron using flags like the following:

* `--exporter.runonce`
* `--exporter.output=/var/lib/node_exporter/textfile_collector/infiniband_exporter.prom`

The collection time of `--collector.switch.rcv-err-details` can take much longer than base metrics due to having to execute `perfquery` once per port.
One way to collect these metrics is collect base metrics with Prometheus scrapes and collect `--collector.switch.rcv-err-details` with runonce using the following flags (example on 8 core system, adjust `--perfquery.max-concurrent` as needed):

* `--exporter.runonce`
* `--exporter.output=/var/lib/node_exporter/textfile_collector/infiniband_exporter.prom`
* `--no-collector.switch.base-metrics`
* `--collector.switch.rcv-err-details`
* `--perfquery.max-concurrent=8`

## Docker

Example of running the Docker container

```
docker run -d -p 9315:9315 \
--name infiniband_exporter \
--cap-add=IPC_LOCK \
--device=/dev/infiniband/umad0 \
treydock/infiniband_exporter
```

## Install

Download the [latest release](https://github.com/treydock/infiniband_exporter/releases)

Add the user that will run `infiniband_exporter`

```
groupadd -r infiniband_exporter
useradd -r -d /var/lib/infiniband_exporter -s /sbin/nologin -M -g infiniband_exporter -M infiniband_exporter
```

Install compiled binaries after extracting tar.gz from release page.

```
cp /tmp/infiniband_exporter /usr/local/bin/infiniband_exporter
```

Add systemd unit file and start service. Modify the `ExecStart` with desired flags.

```
cp systemd/infiniband_exporter.service /etc/systemd/system/infiniband_exporter.service
systemctl daemon-reload
systemctl start infiniband_exporter
```

## Build from source

To produce the `infiniband_exporter` binary:

```
make build
```

Or

```
go get github.com/treydock/infiniband_exporter
```
