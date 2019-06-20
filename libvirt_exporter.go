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
	"log"
	"net/http"
	"os"

	"github.com/libvirt/libvirt-go"
	"github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/kumina/libvirt_exporter/libvirt_schema"
)

// LibvirtExporter implements a Prometheus exporter for libvirt state.
type LibvirtExporter struct {
	uri                string
	exportNovaMetadata bool

	libvirtUpDesc *prometheus.Desc

	libvirtDomainInfoMaxMemDesc    *prometheus.Desc
	libvirtDomainInfoMemoryDesc    *prometheus.Desc
	libvirtDomainInfoNrVirtCpuDesc *prometheus.Desc
	libvirtDomainInfoCpuTimeDesc   *prometheus.Desc

	libvirtDomainBlockRdBytesDesc         *prometheus.Desc
	libvirtDomainBlockRdReqDesc           *prometheus.Desc
	libvirtDomainBlockRdTotalTimesDesc    *prometheus.Desc
	libvirtDomainBlockWrBytesDesc         *prometheus.Desc
	libvirtDomainBlockWrReqDesc           *prometheus.Desc
	libvirtDomainBlockWrTotalTimesDesc    *prometheus.Desc
	libvirtDomainBlockFlushReqDesc        *prometheus.Desc
	libvirtDomainBlockFlushTotalTimesDesc *prometheus.Desc

	libvirtDomainInterfaceRxBytesDesc   *prometheus.Desc
	libvirtDomainInterfaceRxPacketsDesc *prometheus.Desc
	libvirtDomainInterfaceRxErrsDesc    *prometheus.Desc
	libvirtDomainInterfaceRxDropDesc    *prometheus.Desc
	libvirtDomainInterfaceTxBytesDesc   *prometheus.Desc
	libvirtDomainInterfaceTxPacketsDesc *prometheus.Desc
	libvirtDomainInterfaceTxErrsDesc    *prometheus.Desc
	libvirtDomainInterfaceTxDropDesc    *prometheus.Desc
}

// NewLibvirtExporter creates a new Prometheus exporter for libvirt.
func NewLibvirtExporter(uri string, exportNovaMetadata bool) (*LibvirtExporter, error) {
	var domainLabels []string
	if exportNovaMetadata {
		domainLabels = []string{"domain", "uuid", "name", "flavor", "project_name"}
	} else {
		domainLabels = []string{"domain", "uuid"}
	}
	return &LibvirtExporter{
		uri:                uri,
		exportNovaMetadata: exportNovaMetadata,
		libvirtUpDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "", "up"),
			"Whether scraping libvirt's metrics was successful.",
			nil,
			nil),
		libvirtDomainInfoMaxMemDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "maximum_memory_bytes"),
			"Maximum allowed memory of the domain, in bytes.",
			domainLabels,
			nil),
		libvirtDomainInfoMemoryDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "memory_usage_bytes"),
			"Memory usage of the domain, in bytes.",
			domainLabels,
			nil),
		libvirtDomainInfoNrVirtCpuDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "virtual_cpus"),
			"Number of virtual CPUs for the domain.",
			domainLabels,
			nil),
		libvirtDomainInfoCpuTimeDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "cpu_time_seconds_total"),
			"Amount of CPU time used by the domain, in seconds.",
			domainLabels,
			nil),
		libvirtDomainBlockRdBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "read_bytes_total"),
			"Number of bytes read from a block device, in bytes.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockRdReqDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "read_requests_total"),
			"Number of read requests from a block device.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockRdTotalTimesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "read_seconds_total"),
			"Amount of time spent reading from a block device, in seconds.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockWrBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "write_bytes_total"),
			"Number of bytes written from a block device, in bytes.",
			append(domainLabels, "source_file", "target_device"),
			nil),

		libvirtDomainBlockWrReqDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "write_requests_total"),
			"Number of write requests from a block device.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockWrTotalTimesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "write_seconds_total"),
			"Amount of time spent writing from a block device, in seconds.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockFlushReqDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "flush_requests_total"),
			"Number of flush requests from a block device.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockFlushTotalTimesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "flush_seconds_total"),
			"Amount of time spent flushing of a block device, in seconds.",
			append(domainLabels, "source_file", "target_device"),
			nil),

		libvirtDomainInterfaceRxBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_bytes_total"),
			"Number of bytes received on a network interface, in bytes.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceRxPacketsDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_packets_total"),
			"Number of packets received on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceRxErrsDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_errors_total"),
			"Number of packet receive errors on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceRxDropDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_drops_total"),
			"Number of packet receive drops on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceTxBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_bytes_total"),
			"Number of bytes transmitted on a network interface, in bytes.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceTxPacketsDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_packets_total"),
			"Number of packets transmitted on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceTxErrsDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_errors_total"),
			"Number of packet transmit errors on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceTxDropDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_drops_total"),
			"Number of packet transmit drops on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
	}, nil
}

// Describe returns metadata for all Prometheus metrics that may be exported.
func (e *LibvirtExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.libvirtUpDesc

	ch <- e.libvirtDomainInfoMaxMemDesc
	ch <- e.libvirtDomainInfoMemoryDesc
	ch <- e.libvirtDomainInfoNrVirtCpuDesc
	ch <- e.libvirtDomainInfoCpuTimeDesc

	ch <- e.libvirtDomainBlockRdBytesDesc
	ch <- e.libvirtDomainBlockRdReqDesc
	ch <- e.libvirtDomainBlockRdTotalTimesDesc
	ch <- e.libvirtDomainBlockWrBytesDesc
	ch <- e.libvirtDomainBlockWrReqDesc
	ch <- e.libvirtDomainBlockWrTotalTimesDesc
	ch <- e.libvirtDomainBlockFlushReqDesc
	ch <- e.libvirtDomainBlockFlushTotalTimesDesc
}

// Collect scrapes Prometheus metrics from libvirt.
func (e *LibvirtExporter) Collect(ch chan<- prometheus.Metric) {
	err := e.CollectFromLibvirt(ch)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(
			e.libvirtUpDesc,
			prometheus.GaugeValue,
			1.0)
	} else {
		log.Printf("Failed to scrape metrics: %s", err)
		ch <- prometheus.MustNewConstMetric(
			e.libvirtUpDesc,
			prometheus.GaugeValue,
			0.0)
	}
}

// CollectFromLibvirt obtains Prometheus metrics from all domains in a
// libvirt setup.
func (e *LibvirtExporter) CollectFromLibvirt(ch chan<- prometheus.Metric) error {
	conn, err := libvirt.NewConnect(e.uri)
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
			err = e.CollectDomain(ch, domain)
			domain.Free()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CollectDomain extracts Prometheus metrics from a libvirt domain.
func (e *LibvirtExporter) CollectDomain(ch chan<- prometheus.Metric, domain *libvirt.Domain) error {
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

	domainName, err := domain.GetName()
	if err != nil {
		return err
	}
	var domainUUID = desc.UUID

	// Extract domain label valuies
	var domainLabelValues []string
	if e.exportNovaMetadata {
		var (
			novaName        = desc.Metadata.NovaInstance.Name
			novaFlavor      = desc.Metadata.NovaInstance.Flavor.Name
			novaProjectName = desc.Metadata.NovaInstance.Owner.ProjectName
		)
		domainLabelValues = []string{domainName, domainUUID, novaName, novaFlavor, novaProjectName}
	} else {
		domainLabelValues = []string{domainName, domainUUID}
	}

	// Report domain info.
	info, err := domain.GetInfo()
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainInfoMaxMemDesc,
		prometheus.GaugeValue,
		float64(info.MaxMem)*1024,
		domainLabelValues...)
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainInfoMemoryDesc,
		prometheus.GaugeValue,
		float64(info.Memory)*1024,
		domainLabelValues...)
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainInfoNrVirtCpuDesc,
		prometheus.GaugeValue,
		float64(info.NrVirtCpu),
		domainLabelValues...)
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainInfoCpuTimeDesc,
		prometheus.CounterValue,
		float64(info.CpuTime)/1e9,
		domainLabelValues...)

	// Report block device statistics.
	for _, disk := range desc.Devices.Disks {
		if disk.Device == "cdrom" || disk.Device == "fd" {
			continue
		}

		blockStats, err := domain.BlockStats(disk.Target.Device)
		if err != nil {
			return err
		}

		if blockStats.RdBytesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockRdBytesDesc,
				prometheus.CounterValue,
				float64(blockStats.RdBytes),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.RdReqSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockRdReqDesc,
				prometheus.CounterValue,
				float64(blockStats.RdReq),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.RdTotalTimesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockRdTotalTimesDesc,
				prometheus.CounterValue,
				float64(blockStats.RdTotalTimes)/1e9,
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.WrBytesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockWrBytesDesc,
				prometheus.CounterValue,
				float64(blockStats.WrBytes),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.WrReqSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockWrReqDesc,
				prometheus.CounterValue,
				float64(blockStats.WrReq),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.WrTotalTimesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockWrTotalTimesDesc,
				prometheus.CounterValue,
				float64(blockStats.WrTotalTimes)/1e9,
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.FlushReqSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockFlushReqDesc,
				prometheus.CounterValue,
				float64(blockStats.FlushReq),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.FlushTotalTimesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockFlushTotalTimesDesc,
				prometheus.CounterValue,
				float64(blockStats.FlushTotalTimes)/1e9,
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
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
				e.libvirtDomainInterfaceRxBytesDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxBytes),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.RxPacketsSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceRxPacketsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxPackets),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.RxErrsSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceRxErrsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxErrs),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.RxDropSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceRxDropDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxDrop),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.TxBytesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceTxBytesDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxBytes),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.TxPacketsSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceTxPacketsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxPackets),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.TxErrsSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceTxErrsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxErrs),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.TxDropSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceTxDropDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxDrop),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
	}

	return nil
}

func main() {
	var (
		app                       = kingpin.New("libvirt_exporter", "Prometheus metrics exporter for libvirt")
		listenAddress             = app.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9177").String()
		metricsPath               = app.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		libvirtURI                = app.Flag("libvirt.uri", "Libvirt URI from which to extract metrics.").Default("qemu:///system").String()
		libvirtExportNovaMetadata = app.Flag("libvirt.export-nova-metadata", "Export OpenStack Nova specific labels from libvirt domain xml").Default("false").Bool()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	exporter, err := NewLibvirtExporter(*libvirtURI, *libvirtExportNovaMetadata)
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
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
