// Copyright 2020 Trey Dockendorf
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

package collectors

import (
	"context"
	"math"
	"strings"
	"sync"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CollectHCA       = kingpin.Flag("collector.hca", "Enable the HCA collector").Default("false").Bool()
	hcaCollectBase   = kingpin.Flag("collector.hca.base-metrics", "Collect base metrics").Default("true").Bool()
	hcaCollectRcvErr = kingpin.Flag("collector.hca.rcv-err-details", "Collect Rcv Error Details").Default("false").Bool()
)

type HCACollector struct {
	devices                      *[]InfinibandDevice
	logger                       log.Logger
	collector                    string
	PortXmitData                 *prometheus.Desc
	PortRcvData                  *prometheus.Desc
	PortXmitPkts                 *prometheus.Desc
	PortRcvPkts                  *prometheus.Desc
	PortUnicastXmitPkts          *prometheus.Desc
	PortUnicastRcvPkts           *prometheus.Desc
	PortMulticastXmitPkts        *prometheus.Desc
	PortMulticastRcvPkts         *prometheus.Desc
	SymbolErrorCounter           *prometheus.Desc
	LinkErrorRecoveryCounter     *prometheus.Desc
	LinkDownedCounter            *prometheus.Desc
	PortRcvErrors                *prometheus.Desc
	PortRcvRemotePhysicalErrors  *prometheus.Desc
	PortRcvSwitchRelayErrors     *prometheus.Desc
	PortXmitDiscards             *prometheus.Desc
	PortXmitConstraintErrors     *prometheus.Desc
	PortRcvConstraintErrors      *prometheus.Desc
	LocalLinkIntegrityErrors     *prometheus.Desc
	ExcessiveBufferOverrunErrors *prometheus.Desc
	VL15Dropped                  *prometheus.Desc
	PortXmitWait                 *prometheus.Desc
	QP1Dropped                   *prometheus.Desc
	PortLocalPhysicalErrors      *prometheus.Desc
	PortMalformedPktErrors       *prometheus.Desc
	PortBufferOverrunErrors      *prometheus.Desc
	PortDLIDMappingErrors        *prometheus.Desc
	PortVLMappingErrors          *prometheus.Desc
	PortLoopingErrors            *prometheus.Desc
	Rate                         *prometheus.Desc
	RawRate                      *prometheus.Desc
	Uplink                       *prometheus.Desc
	Info                         *prometheus.Desc
}

func NewHCACollector(devices *[]InfinibandDevice, runonce bool, logger log.Logger) *HCACollector {
	labels := []string{"guid", "port"}
	collector := "hca"
	if runonce {
		collector = "hca-runonce"
	}
	return &HCACollector{
		devices:   devices,
		logger:    log.With(logger, "collector", collector),
		collector: collector,
		PortXmitData: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_transmit_data_bytes_total"),
			"Infiniband HCA port PortXmitData", labels, nil),
		PortRcvData: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_receive_data_bytes_total"),
			"Infiniband HCA port PortRcvData", labels, nil),
		PortXmitPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_transmit_packets_total"),
			"Infiniband HCA port PortXmitPkts", labels, nil),
		PortRcvPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_receive_packets_total"),
			"Infiniband HCA port PortRcvPkts", labels, nil),
		PortUnicastXmitPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_unicast_transmit_packets_total"),
			"Infiniband HCA port PortUnicastXmitPkts", labels, nil),
		PortUnicastRcvPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_unicast_receive_packets_total"),
			"Infiniband HCA port PortUnicastRcvPkts", labels, nil),
		PortMulticastXmitPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_multicast_transmit_packets_total"),
			"Infiniband HCA port PortMulticastXmitPkts", labels, nil),
		PortMulticastRcvPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_multicast_receive_packets_total"),
			"Infiniband HCA port PortMulticastRcvPkts", labels, nil),
		SymbolErrorCounter: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_symbol_error_total"),
			"Infiniband HCA port SymbolErrorCounter", labels, nil),
		LinkErrorRecoveryCounter: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_link_error_recovery_total"),
			"Infiniband HCA port LinkErrorRecoveryCounter", labels, nil),
		LinkDownedCounter: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_link_downed_total"),
			"Infiniband HCA port LinkDownedCounter", labels, nil),
		PortRcvErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_receive_errors_total"),
			"Infiniband HCA port PortRcvErrors", labels, nil),
		PortRcvRemotePhysicalErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_receive_remote_physical_errors_total"),
			"Infiniband HCA port PortRcvRemotePhysicalErrors", labels, nil),
		PortRcvSwitchRelayErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_receive_switch_relay_errors_total"),
			"Infiniband HCA port PortRcvSwitchRelayErrors", labels, nil),
		PortXmitDiscards: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_transmit_discards_total"),
			"Infiniband HCA port PortXmitDiscards", labels, nil),
		PortXmitConstraintErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_transmit_constraint_errors_total"),
			"Infiniband HCA port PortXmitConstraintErrors", labels, nil),
		PortRcvConstraintErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_receive_constraint_errors_total"),
			"Infiniband HCA port PortRcvConstraintErrors", labels, nil),
		LocalLinkIntegrityErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_local_link_integrity_errors_total"),
			"Infiniband HCA port LocalLinkIntegrityErrors", labels, nil),
		ExcessiveBufferOverrunErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_excessive_buffer_overrun_errors_total"),
			"Infiniband HCA port ExcessiveBufferOverrunErrors", labels, nil),
		VL15Dropped: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_vl15_dropped_total"),
			"Infiniband HCA port VL15Dropped", labels, nil),
		PortXmitWait: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_transmit_wait_total"),
			"Infiniband HCA port PortXmitWait", labels, nil),
		QP1Dropped: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_qp1_dropped_total"),
			"Infiniband HCA port QP1Dropped", labels, nil),
		PortLocalPhysicalErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_local_physical_errors_total"),
			"Infiniband HCA port PortLocalPhysicalErrors", labels, nil),
		PortMalformedPktErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_malformed_packet_errors_total"),
			"Infiniband HCA port PortMalformedPktErrors", labels, nil),
		PortBufferOverrunErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_buffer_overrun_errors_total"),
			"Infiniband HCA port PortBufferOverrunErrors", labels, nil),
		PortDLIDMappingErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_dli_mapping_errors_total"),
			"Infiniband HCA port PortDLIDMappingErrors", labels, nil),
		PortVLMappingErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_vl_mapping_errors_total"),
			"Infiniband HCA port PortVLMappingErrors", labels, nil),
		PortLoopingErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "port_looping_errors_total"),
			"Infiniband HCA port PortLoopingErrors", labels, nil),
		Rate: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "rate_bytes_per_second"),
			"Infiniband HCA rate", []string{"guid"}, nil),
		RawRate: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "raw_rate_bytes_per_second"),
			"Infiniband HCA raw rate", []string{"guid"}, nil),
		Uplink: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "uplink_info"),
			"Infiniband HCA uplink information", append(labels, []string{"hca", "uplink", "uplink_guid", "uplink_type", "uplink_port", "uplink_port_name", "uplink_lid"}...), nil),
		Info: prometheus.NewDesc(prometheus.BuildFQName(namespace, "hca", "info"),
			"Infiniband HCA information", []string{"guid", "hca", "port_name", "lid"}, nil),
	}
}

func (h *HCACollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- h.PortXmitData
	ch <- h.PortRcvData
	ch <- h.PortXmitPkts
	ch <- h.PortRcvPkts
	ch <- h.PortUnicastXmitPkts
	ch <- h.PortUnicastRcvPkts
	ch <- h.PortMulticastXmitPkts
	ch <- h.PortMulticastRcvPkts
	ch <- h.SymbolErrorCounter
	ch <- h.LinkErrorRecoveryCounter
	ch <- h.LinkDownedCounter
	ch <- h.PortRcvErrors
	ch <- h.PortRcvRemotePhysicalErrors
	ch <- h.PortRcvSwitchRelayErrors
	ch <- h.PortXmitDiscards
	ch <- h.PortXmitConstraintErrors
	ch <- h.PortRcvConstraintErrors
	ch <- h.LocalLinkIntegrityErrors
	ch <- h.ExcessiveBufferOverrunErrors
	ch <- h.VL15Dropped
	ch <- h.PortXmitWait
	ch <- h.QP1Dropped
	ch <- h.PortLocalPhysicalErrors
	ch <- h.PortMalformedPktErrors
	ch <- h.PortBufferOverrunErrors
	ch <- h.PortDLIDMappingErrors
	ch <- h.PortVLMappingErrors
	ch <- h.PortLoopingErrors
	ch <- h.Rate
	ch <- h.RawRate
	ch <- h.Uplink
	ch <- h.Info
}

func (h *HCACollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()
	counters, errors, timeouts := h.collect()
	for _, c := range counters {
		if !math.IsNaN(c.PortXmitData) {
			ch <- prometheus.MustNewConstMetric(h.PortXmitData, prometheus.CounterValue, c.PortXmitData, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortRcvData) {
			ch <- prometheus.MustNewConstMetric(h.PortRcvData, prometheus.CounterValue, c.PortRcvData, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortXmitPkts) {
			ch <- prometheus.MustNewConstMetric(h.PortXmitPkts, prometheus.CounterValue, c.PortXmitPkts, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortRcvPkts) {
			ch <- prometheus.MustNewConstMetric(h.PortRcvPkts, prometheus.CounterValue, c.PortRcvPkts, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortUnicastXmitPkts) {
			ch <- prometheus.MustNewConstMetric(h.PortUnicastXmitPkts, prometheus.CounterValue, c.PortUnicastXmitPkts, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortUnicastRcvPkts) {
			ch <- prometheus.MustNewConstMetric(h.PortUnicastRcvPkts, prometheus.CounterValue, c.PortUnicastRcvPkts, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortMulticastXmitPkts) {
			ch <- prometheus.MustNewConstMetric(h.PortMulticastXmitPkts, prometheus.CounterValue, c.PortMulticastXmitPkts, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortMulticastRcvPkts) {
			ch <- prometheus.MustNewConstMetric(h.PortMulticastRcvPkts, prometheus.CounterValue, c.PortMulticastRcvPkts, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.SymbolErrorCounter) {
			ch <- prometheus.MustNewConstMetric(h.SymbolErrorCounter, prometheus.CounterValue, c.SymbolErrorCounter, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.LinkErrorRecoveryCounter) {
			ch <- prometheus.MustNewConstMetric(h.LinkErrorRecoveryCounter, prometheus.CounterValue, c.LinkErrorRecoveryCounter, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.LinkDownedCounter) {
			ch <- prometheus.MustNewConstMetric(h.LinkDownedCounter, prometheus.CounterValue, c.LinkDownedCounter, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortRcvErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortRcvErrors, prometheus.CounterValue, c.PortRcvErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortRcvRemotePhysicalErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortRcvRemotePhysicalErrors, prometheus.CounterValue, c.PortRcvRemotePhysicalErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortRcvSwitchRelayErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortRcvSwitchRelayErrors, prometheus.CounterValue, c.PortRcvSwitchRelayErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortXmitDiscards) {
			ch <- prometheus.MustNewConstMetric(h.PortXmitDiscards, prometheus.CounterValue, c.PortXmitDiscards, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortXmitConstraintErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortXmitConstraintErrors, prometheus.CounterValue, c.PortXmitConstraintErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortRcvConstraintErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortRcvConstraintErrors, prometheus.CounterValue, c.PortRcvConstraintErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.LocalLinkIntegrityErrors) {
			ch <- prometheus.MustNewConstMetric(h.LocalLinkIntegrityErrors, prometheus.CounterValue, c.LocalLinkIntegrityErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.ExcessiveBufferOverrunErrors) {
			ch <- prometheus.MustNewConstMetric(h.ExcessiveBufferOverrunErrors, prometheus.CounterValue, c.ExcessiveBufferOverrunErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.VL15Dropped) {
			ch <- prometheus.MustNewConstMetric(h.VL15Dropped, prometheus.CounterValue, c.VL15Dropped, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortXmitWait) {
			ch <- prometheus.MustNewConstMetric(h.PortXmitWait, prometheus.CounterValue, c.PortXmitWait, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.QP1Dropped) {
			ch <- prometheus.MustNewConstMetric(h.QP1Dropped, prometheus.CounterValue, c.QP1Dropped, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortLocalPhysicalErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortLocalPhysicalErrors, prometheus.CounterValue, c.PortLocalPhysicalErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortMalformedPktErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortMalformedPktErrors, prometheus.CounterValue, c.PortMalformedPktErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortBufferOverrunErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortBufferOverrunErrors, prometheus.CounterValue, c.PortBufferOverrunErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortDLIDMappingErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortDLIDMappingErrors, prometheus.CounterValue, c.PortDLIDMappingErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortVLMappingErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortVLMappingErrors, prometheus.CounterValue, c.PortVLMappingErrors, c.device.GUID, c.PortSelect)
		}
		if !math.IsNaN(c.PortLoopingErrors) {
			ch <- prometheus.MustNewConstMetric(h.PortLoopingErrors, prometheus.CounterValue, c.PortLoopingErrors, c.device.GUID, c.PortSelect)
		}
	}
	if *hcaCollectBase {
		for _, device := range *h.devices {
			ch <- prometheus.MustNewConstMetric(h.Rate, prometheus.GaugeValue, device.Rate, device.GUID)
			ch <- prometheus.MustNewConstMetric(h.RawRate, prometheus.GaugeValue, device.RawRate, device.GUID)
			ch <- prometheus.MustNewConstMetric(h.Info, prometheus.GaugeValue, 1, device.GUID, device.Name, device.PortName, device.LID)
			for port, uplink := range device.Uplinks {
				ch <- prometheus.MustNewConstMetric(h.Uplink, prometheus.GaugeValue, 1, device.GUID, port, device.Name, uplink.Name, uplink.GUID, uplink.Type, uplink.PortNumber, uplink.PortName, uplink.LID)
			}
		}
	}
	ch <- prometheus.MustNewConstMetric(collectErrors, prometheus.GaugeValue, errors, h.collector)
	ch <- prometheus.MustNewConstMetric(collecTimeouts, prometheus.GaugeValue, timeouts, h.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), h.collector)
	if strings.HasSuffix(h.collector, "-runonce") {
		ch <- prometheus.MustNewConstMetric(lastExecution, prometheus.GaugeValue, float64(time.Now().Unix()), h.collector)
	}
}

func (h *HCACollector) collect() ([]PerfQueryCounters, float64, float64) {
	var counters []PerfQueryCounters
	var countersLock sync.Mutex
	var errors, timeouts float64
	limit := make(chan int, *maxConcurrent)
	wg := &sync.WaitGroup{}
	for _, device := range *h.devices {
		limit <- 1
		wg.Add(1)
		go func(device InfinibandDevice) {
			defer func() {
				<-limit
				wg.Done()
			}()
			ctxExtended, cancelExtended := context.WithTimeout(context.Background(), *perfqueryTimeout)
			defer cancelExtended()
			ports := getDevicePorts(device.Uplinks)
			perfqueryPorts := strings.Join(ports, ",")
			extendedOut, err := PerfqueryExec(device.GUID, perfqueryPorts, []string{"-l", "-x"}, ctxExtended)
			if err == context.DeadlineExceeded {
				level.Error(h.logger).Log("msg", "Timeout collecting extended perfquery counters", "guid", device.GUID)
				timeouts++
			} else if err != nil {
				level.Error(h.logger).Log("msg", "Error collecting extended perfquery counters", "guid", device.GUID)
				errors++
			}
			if err != nil {
				return
			}
			deviceCounters, errs := perfqueryParse(device, extendedOut, h.logger)
			errors = errors + errs
			if *hcaCollectBase {
				level.Debug(h.logger).Log("msg", "Adding parsed counters", "count", len(deviceCounters), "guid", device.GUID, "name", device.Name)
				countersLock.Lock()
				counters = append(counters, deviceCounters...)
				countersLock.Unlock()
			}
			if *hcaCollectRcvErr {
				for _, deviceCounter := range deviceCounters {
					ctxRcvErr, cancelRcvErr := context.WithTimeout(context.Background(), *perfqueryTimeout)
					defer cancelRcvErr()
					rcvErrOut, err := PerfqueryExec(device.GUID, deviceCounter.PortSelect, []string{"-E"}, ctxRcvErr)
					if err == context.DeadlineExceeded {
						level.Error(h.logger).Log("msg", "Timeout collecting rcvErr perfquery counters", "guid", device.GUID)
						timeouts++
						continue
					} else if err != nil {
						level.Error(h.logger).Log("msg", "Error collecting rcvErr perfquery counters", "guid", device.GUID)
						errors++
						continue
					}
					rcvErrCounters, errs := perfqueryParse(device, rcvErrOut, h.logger)
					errors = errors + errs
					countersLock.Lock()
					counters = append(counters, rcvErrCounters...)
					countersLock.Unlock()
				}
			}
		}(device)
	}
	wg.Wait()
	close(limit)
	return counters, errors, timeouts
}
