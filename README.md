# Prometheus libvirt exporter

This repository provides code for a Prometheus metrics exporter
for [libvirt](https://libvirt.org/). This exporter connects to any
libvirt daemon and exports per-domain metrics related to CPU, memory,
disk and network usage. By default, this exporter listens on TCP port
9177.

This exporter makes use of
[libvirt-go](https://github.com/libvirt/libvirt-go), the official Go
bindings for libvirt. Ideally, this exporter should make use of the
`GetAllDomainStats()` API call to extract all relevant metrics.
Unfortunately, we at Kumina still need this exporter to be compatible
with older versions of libvirt that don't support this API call.

The following metrics/labels are being exported:

```
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

At Kumina we want to perform a single build of this exporter, deploying
it to a variety of Linux distribution versions. This is why this
repository contains a shell script, `build_static.sh`, that builds a
statically linked copy of this exporter in an Alpine Linux based
container.
