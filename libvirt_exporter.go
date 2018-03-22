// Copyright 2017 Kumina, https://kumina.nl/
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/xml"
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	libvirt "github.com/libvirt/libvirt-go"

	"github.com/kumina/libvirt_exporter/libvirt_schema"
)

/*
#cgo LDFLAGS: -lvirt
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
*/
import "C"

const (
	VIR_DOMAIN_MEMORY_STAT_SWAP_IN        = C.VIR_DOMAIN_MEMORY_STAT_SWAP_IN
	VIR_DOMAIN_MEMORY_STAT_SWAP_OUT       = C.VIR_DOMAIN_MEMORY_STAT_SWAP_OUT
	VIR_DOMAIN_MEMORY_STAT_MAJOR_FAULT    = C.VIR_DOMAIN_MEMORY_STAT_MAJOR_FAULT
	VIR_DOMAIN_MEMORY_STAT_MINOR_FAULT    = C.VIR_DOMAIN_MEMORY_STAT_MINOR_FAULT
	VIR_DOMAIN_MEMORY_STAT_UNUSED         = C.VIR_DOMAIN_MEMORY_STAT_UNUSED
	VIR_DOMAIN_MEMORY_STAT_AVAILABLE      = C.VIR_DOMAIN_MEMORY_STAT_AVAILABLE
	VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON = C.VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON
	VIR_DOMAIN_MEMORY_STAT_RSS            = C.VIR_DOMAIN_MEMORY_STAT_RSS
	VIR_DOMAIN_MEMORY_STAT_NR             = C.VIR_DOMAIN_MEMORY_STAT_NR

	VIR_DOMAIN_CPU_STATS_CPUTIME          = C.VIR_DOMAIN_CPU_STATS_CPUTIME
	VIR_DOMAIN_CPU_STATS_USERTIME         = C.VIR_DOMAIN_CPU_STATS_USERTIME
	VIR_DOMAIN_CPU_STATS_SYSTEMTIME       = C.VIR_DOMAIN_CPU_STATS_SYSTEMTIME

	libvirtdAddress = "qemu:///system"
)

var (
	libvirtUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "", "up"),
		"Whether scraping libvirt's metrics was successful.",
		nil,
		nil)

	libvirtDomainInfoMaxMemDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "maximum_memory_bytes"),
		"Maximum allowed memory of the domain, in bytes.",
		[]string{"domain"},
		nil)
	libvirtDomainInfoMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "memory_usage_bytes"),
		"Memory usage of the domain, in bytes.",
		[]string{"domain"},
		nil)
	libvirtDomainInfoNrVirtCpuDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "virtual_cpus"),
		"Number of virtual CPUs for the domain.",
		[]string{"domain"},
		nil)
	libvirtDomainInfoCpuTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "cpu_time_seconds_total"),
		"Amount of CPU time used by the domain, in seconds.",
		[]string{"domain"},
		nil)

	libvirtDomainBlockRdBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "read_bytes_total"),
		"Number of bytes read from a block device, in bytes.",
		[]string{"domain", "source_file", "target_device"},
		nil)
	libvirtDomainBlockRdReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "read_requests_total"),
		"Number of read requests from a block device.",
		[]string{"domain", "source_file", "target_device"},
		nil)
	libvirtDomainBlockRdTotalTimesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "read_seconds_total"),
		"Amount of time spent reading from a block device, in seconds.",
		[]string{"domain", "source_file", "target_device"},
		nil)
	libvirtDomainBlockWrBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "write_bytes_total"),
		"Number of bytes written from a block device, in bytes.",
		[]string{"domain", "source_file", "target_device"},
		nil)
	libvirtDomainBlockWrReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "write_requests_total"),
		"Number of write requests from a block device.",
		[]string{"domain", "source_file", "target_device"},
		nil)
	libvirtDomainBlockWrTotalTimesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "write_seconds_total"),
		"Amount of time spent writing from a block device, in seconds.",
		[]string{"domain", "source_file", "target_device"},
		nil)
	libvirtDomainBlockFlushReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "flush_requests_total"),
		"Number of flush requests from a block device.",
		[]string{"domain", "source_file", "target_device"},
		nil)
	libvirtDomainBlockFlushTotalTimesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "flush_seconds_total"),
		"Amount of time spent flushing of a block device, in seconds.",
		[]string{"domain", "source_file", "target_device"},
		nil)

	libvirtDomainInterfaceRxBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_bytes_total"),
		"Number of bytes received on a network interface, in bytes.",
		[]string{"domain", "source_bridge", "target_device"},
		nil)
	libvirtDomainInterfaceRxPacketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_packets_total"),
		"Number of packets received on a network interface.",
		[]string{"domain", "source_bridge", "target_device"},
		nil)
	libvirtDomainInterfaceRxErrsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_errors_total"),
		"Number of packet receive errors on a network interface.",
		[]string{"domain", "source_bridge", "target_device"},
		nil)
	libvirtDomainInterfaceRxDropDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_drops_total"),
		"Number of packet receive drops on a network interface.",
		[]string{"domain", "source_bridge", "target_device"},
		nil)
	libvirtDomainInterfaceTxBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_bytes_total"),
		"Number of bytes transmitted on a network interface, in bytes.",
		[]string{"domain", "source_bridge", "target_device"},
		nil)
	libvirtDomainInterfaceTxPacketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_packets_total"),
		"Number of packets transmitted on a network interface.",
		[]string{"domain", "source_bridge", "target_device"},
		nil)
	libvirtDomainInterfaceTxErrsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_errors_total"),
		"Number of packet transmit errors on a network interface.",
		[]string{"domain", "source_bridge", "target_device"},
		nil)
	libvirtDomainInterfaceTxDropDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_drops_total"),
		"Number of packet transmit drops on a network interface.",
		[]string{"domain", "source_bridge", "target_device"},
		nil)
)

// CollectDomain extracts Prometheus metrics from a libvirt domain.
func CollectDomain(ch chan<- prometheus.Metric, domain *libvirt.Domain) error {
	domainName, err := domain.GetName()
	if err != nil {
		return err
	}

	// Decode XML description of domain to get block device names, etc.
	xmlDesc, err := domain.GetXMLDesc(0)
	if err != nil {
		return err
	}
	var desc libvirt_schema.Domain
	err = xml.Unmarshal([]byte(xmlDesc), &desc)
	if err != nil {
		return err
	}

	// Report domain info.
	info, err := domain.GetInfo()
	if err != nil {
		return err
	}
	memstats, err := domain.MemoryStats(VIR_DOMAIN_MEMORY_STAT_NR, 0)
	if err != nil {
		return err
	}
	var mem_unused uint64 = 0
	mem_avail := info.MaxMem
	for _, stat := range memstats {
		switch stat.Tag {
		case VIR_DOMAIN_MEMORY_STAT_UNUSED:
			mem_unused = stat.Val
		case VIR_DOMAIN_MEMORY_STAT_AVAILABLE:
			mem_avail = stat.Val
		}
	}
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoMaxMemDesc,
		prometheus.GaugeValue,
		float64(info.MaxMem)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoMemoryDesc,
		prometheus.GaugeValue,
		float64(mem_avail - mem_unused)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoNrVirtCpuDesc,
		prometheus.GaugeValue,
		float64(info.NrVirtCpu),
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoCpuTimeDesc,
		prometheus.CounterValue,
		float64(info.CpuTime)/1e9,
		domainName)

	// Report block device statistics.
	for _, disk := range desc.Devices.Disks {
		blockStats, err := domain.BlockStats(disk.Target.Device)
		if err != nil {
			return err
		}

		if blockStats.RdBytesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockRdBytesDesc,
				prometheus.CounterValue,
				float64(blockStats.RdBytes),
				domainName,
				disk.Source.File,
				disk.Target.Device)
		}
		if blockStats.RdReqSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockRdReqDesc,
				prometheus.CounterValue,
				float64(blockStats.RdReq),
				domainName,
				disk.Source.File,
				disk.Target.Device)
		}
		if blockStats.RdTotalTimesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockRdTotalTimesDesc,
				prometheus.CounterValue,
				float64(blockStats.RdTotalTimes)/1e9,
				domainName,
				disk.Source.File,
				disk.Target.Device)
		}
		if blockStats.WrBytesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockWrBytesDesc,
				prometheus.CounterValue,
				float64(blockStats.WrBytes),
				domainName,
				disk.Source.File,
				disk.Target.Device)
		}
		if blockStats.WrReqSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockWrReqDesc,
				prometheus.CounterValue,
				float64(blockStats.WrReq),
				domainName,
				disk.Source.File,
				disk.Target.Device)
		}
		if blockStats.WrTotalTimesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockWrTotalTimesDesc,
				prometheus.CounterValue,
				float64(blockStats.WrTotalTimes)/1e9,
				domainName,
				disk.Source.File,
				disk.Target.Device)
		}
		if blockStats.FlushReqSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockFlushReqDesc,
				prometheus.CounterValue,
				float64(blockStats.FlushReq),
				domainName,
				disk.Source.File,
				disk.Target.Device)
		}
		if blockStats.FlushTotalTimesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockFlushTotalTimesDesc,
				prometheus.CounterValue,
				float64(blockStats.FlushTotalTimes)/1e9,
				domainName,
				disk.Source.File,
				disk.Target.Device)
		}
		// Skip "Errs", as the documentation does not clearly
		// explain what this means.
	}

	// Report network interface statistics.
	for _, iface := range desc.Devices.Interfaces {
		if iface.Target.Device == "" {
			continue
		}
		interfaceStats, err := domain.InterfaceStats(iface.Target.Device)
		if err != nil {
			return err
		}

		if interfaceStats.RxBytesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceRxBytesDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxBytes),
				domainName,
				iface.Source.Bridge,
				iface.Target.Device)
		}
		if interfaceStats.RxPacketsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceRxPacketsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxPackets),
				domainName,
				iface.Source.Bridge,
				iface.Target.Device)
		}
		if interfaceStats.RxErrsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceRxErrsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxErrs),
				domainName,
				iface.Source.Bridge,
				iface.Target.Device)
		}
		if interfaceStats.RxDropSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceRxDropDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxDrop),
				domainName,
				iface.Source.Bridge,
				iface.Target.Device)
		}
		if interfaceStats.TxBytesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceTxBytesDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxBytes),
				domainName,
				iface.Source.Bridge,
				iface.Target.Device)
		}
		if interfaceStats.TxPacketsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceTxPacketsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxPackets),
				domainName,
				iface.Source.Bridge,
				iface.Target.Device)
		}
		if interfaceStats.TxErrsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceTxErrsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxErrs),
				domainName,
				iface.Source.Bridge,
				iface.Target.Device)
		}
		if interfaceStats.TxDropSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceTxDropDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxDrop),
				domainName,
				iface.Source.Bridge,
				iface.Target.Device)
		}
	}

	return nil
}

// CollectFromLibvirt obtains Prometheus metrics from all domains in a
// libvirt setup.
func CollectFromLibvirt(ch chan<- prometheus.Metric, uri string) error {
	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Use ListDomains() as opposed to using ListAllDomains(), as
	// the latter is unsupported when talking to a system using
	// libvirt 0.9.12 or older.
	domainIds, err := conn.ListDomains()
	if err != nil {
		return err
	}
	for _, id := range domainIds {
		domain, err := conn.LookupDomainById(id)
		if err == nil {
			err = CollectDomain(ch, domain)
			domain.Free()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// LibvirtExporter implements a Prometheus exporter for libvirt state.
type LibvirtExporter struct {
	uri string
}

// NewLibvirtExporter creates a new Prometheus exporter for libvirt.
func NewLibvirtExporter(uri string) (*LibvirtExporter, error) {
	return &LibvirtExporter{
		uri: uri,
	}, nil
}

// Describe returns metadata for all Prometheus metrics that may be exported.
func (e *LibvirtExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- libvirtUpDesc

	ch <- libvirtDomainInfoMaxMemDesc
	ch <- libvirtDomainInfoMemoryDesc
	ch <- libvirtDomainInfoNrVirtCpuDesc
	ch <- libvirtDomainInfoCpuTimeDesc

	ch <- libvirtDomainBlockRdBytesDesc
	ch <- libvirtDomainBlockRdReqDesc
	ch <- libvirtDomainBlockRdTotalTimesDesc
	ch <- libvirtDomainBlockWrBytesDesc
	ch <- libvirtDomainBlockWrReqDesc
	ch <- libvirtDomainBlockWrTotalTimesDesc
	ch <- libvirtDomainBlockFlushReqDesc
	ch <- libvirtDomainBlockFlushTotalTimesDesc
}

// Collect scrapes Prometheus metrics from libvirt.
func (e *LibvirtExporter) Collect(ch chan<- prometheus.Metric) {
	err := CollectFromLibvirt(ch, e.uri)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(
			libvirtUpDesc,
			prometheus.GaugeValue,
			1.0)
	} else {
		log.Printf("Failed to scrape metrics: %s", err)
		ch <- prometheus.MustNewConstMetric(
			libvirtUpDesc,
			prometheus.GaugeValue,
			0.0)
	}
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9177", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		libvirtURI    = flag.String("libvirt.uri", "qemu:///system", "Libvirt URI from which to extract metrics.")
	)
	flag.Parse()

	exporter, err := NewLibvirtExporter(*libvirtURI)
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>Libvirt Exporter</title></head>
			<body>
			<h1>Libvirt Exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
