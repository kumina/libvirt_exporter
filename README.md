# Prometheus libvirt exporter

## this is fork repo

this repo fork [kumina-libvirt_exporter](https://github.com/kumina/libvirt_exporter)

## libvirt api

The libvirt api is [here](https://libvirt.org/html/)

## libvirt go package

This exporter makes use of
[libvirt-go](https://github.com/libvirt/libvirt-go), the official Go
bindings for libvirt. Ideally, this exporter should make use of the
`GetAllDomainStats()` API call to extract all relevant metrics.
Unfortunately, we at Kumina still need this exporter to be compatible
with older versions of libvirt that don't support this API call.

## export metrics

### base in forks repo
The following metrics/labels are being exported:

```bash
libvirt_domain_block_stats_read_bytes_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_read_requests_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_write_bytes_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_write_requests_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_info_cpu_time_seconds_total{domain="..."}
libvirt_domain_info_maximum_memory_bytes{domain="..."}
libvirt_domain_info_memory_usage_bytes{domain="..."}
libvirt_domain_info_virtual_cpus{domain="..."}
libvirt_domain_interface_stats_receive_bytes_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_receive_drops_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_receive_errors_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_receive_packets_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_bytes_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_drops_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_errors_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_packets_total{domain="...",source_bridge="...",target_device="..."}
libvirt_up
```

### what i do

add block info metrics

```bash
# HELP libvirt_domain_block_info_allocation host physical size in bytes of the image container, in bytes.
libvirt_domain_block_info_allocation{domain="...",source_file="",target_device="..."} 
# HELP libvirt_domain_block_info_capacity how much storage the guest will see, in bytes.
libvirt_domain_block_info_capacity{domain="...",source_file="",target_device="..."} 
# HELP libvirt_domain_block_info_physical host storage in bytes occupied by the image, in bytes.
libvirt_domain_block_info_physical{domain="...",source_file="",target_device="..."} 
```

## use docker build 
At Kumina we want to perform a single build of this exporter, deploying
it to a variety of Linux distribution versions. This is why this
repository contains a shell script, `build_static.sh`, that builds a
statically linked copy of this exporter in an Alpine Linux based
container.

## how to use

This repository provides code for a Prometheus metrics exporter
for [libvirt](https://libvirt.org/). This exporter connects to any
libvirt daemon and exports per-domain metrics related to CPU, memory,
disk and network usage. By default, this exporter listens on TCP port
9177.

after build ，you can to see hlep

```bash
usage: libvirt_exporter [<flags>]

Prometheus metrics exporter for libvirt

Flags:
  --help                        Show context-sensitive help (also try --help-long and --help-man).
  --web.listen-address=":9177"  Address to listen on for web interface and telemetry.
  --web.telemetry-path="/metrics"
                                Path under which to expose metrics.
  --libvirt.uri="qemu:///system"
                                Libvirt URI from which to extract metrics.

```

use expamply

```bash
./libvirt_exporter --libvirt.uri=qemu+unix:///system?socket=/var/run/libvirt/libvirt-sock
```

the libvirt remote connect config is [here](https://libvirt.org/remote.html)
