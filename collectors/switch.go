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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	CollectSwitch       = kingpin.Flag("collector.switch", "Enable the switch collector").Default("true").Bool()
	switchCollectRcvErr = kingpin.Flag("collector.switch.rcv-err-details", "Collect Rcv Error Details").Default("false").Bool()
)

type SwitchCollector struct {
	switches                     *[]InfinibandDevice
	logger                       log.Logger
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
}

func NewSwitchCollector(switches *[]InfinibandDevice, logger log.Logger) *SwitchCollector {
	labels := []string{"guid", "port", "switch"}
	return &SwitchCollector{
		switches: switches,
		logger:   log.With(logger, "collector", "switch"),
		PortXmitData: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_transmit_data_bytes_total"),
			"Infiniband switch port PortXmitData", labels, nil),
		PortRcvData: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_receive_data_bytes_total"),
			"Infiniband switch port PortRcvData", labels, nil),
		PortXmitPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_transmit_packets_total"),
			"Infiniband switch port PortXmitPkts", labels, nil),
		PortRcvPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_receive_packets_total"),
			"Infiniband switch port PortRcvPkts", labels, nil),
		PortUnicastXmitPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_unicast_transmit_packets_total"),
			"Infiniband switch port PortUnicastXmitPkts", labels, nil),
		PortUnicastRcvPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_unicast_receive_packets_total"),
			"Infiniband switch port PortUnicastRcvPkts", labels, nil),
		PortMulticastXmitPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_multicast_transmit_packets_total"),
			"Infiniband switch port PortMulticastXmitPkts", labels, nil),
		PortMulticastRcvPkts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_multicast_receive_packets_total"),
			"Infiniband switch port PortMulticastRcvPkts", labels, nil),
		SymbolErrorCounter: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_symbol_error_total"),
			"Infiniband switch port SymbolErrorCounter", labels, nil),
		LinkErrorRecoveryCounter: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_link_error_recovery_total"),
			"Infiniband switch port LinkErrorRecoveryCounter", labels, nil),
		LinkDownedCounter: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_link_downed_total"),
			"Infiniband switch port LinkDownedCounter", labels, nil),
		PortRcvErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_receive_errors_total"),
			"Infiniband switch port PortRcvErrors", labels, nil),
		PortRcvRemotePhysicalErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_receive_remote_physical_errors_total"),
			"Infiniband switch port PortRcvRemotePhysicalErrors", labels, nil),
		PortRcvSwitchRelayErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_receive_switch_relay_errors_total"),
			"Infiniband switch port PortRcvSwitchRelayErrors", labels, nil),
		PortXmitDiscards: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_transmit_discards_total"),
			"Infiniband switch port PortXmitDiscards", labels, nil),
		PortXmitConstraintErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_transmit_constraint_errors_total"),
			"Infiniband switch port PortXmitConstraintErrors", labels, nil),
		PortRcvConstraintErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_receive_constraint_errors_total"),
			"Infiniband switch port PortRcvConstraintErrors", labels, nil),
		LocalLinkIntegrityErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_local_link_integrity_errors_total"),
			"Infiniband switch port LocalLinkIntegrityErrors", labels, nil),
		ExcessiveBufferOverrunErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_excessive_buffer_overrun_errors_total"),
			"Infiniband switch port ExcessiveBufferOverrunErrors", labels, nil),
		VL15Dropped: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_vl15_dropped_total"),
			"Infiniband switch port VL15Dropped", labels, nil),
		PortXmitWait: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_transmit_wait_total"),
			"Infiniband switch port PortXmitWait", labels, nil),
		QP1Dropped: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_qp1_dropped_total"),
			"Infiniband switch port QP1Dropped", labels, nil),
		PortLocalPhysicalErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_local_physical_errors_total"),
			"Infiniband switch port PortLocalPhysicalErrors", labels, nil),
		PortMalformedPktErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_malformed_packet_errors_total"),
			"Infiniband switch port PortMalformedPktErrors", labels, nil),
		PortBufferOverrunErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_buffer_overrun_errors_total"),
			"Infiniband switch port PortBufferOverrunErrors", labels, nil),
		PortDLIDMappingErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_dli_mapping_errors_total"),
			"Infiniband switch port PortDLIDMappingErrors", labels, nil),
		PortVLMappingErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_vl_mapping_errors_total"),
			"Infiniband switch port PortVLMappingErrors", labels, nil),
		PortLoopingErrors: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "port_looping_errors_total"),
			"Infiniband switch port PortLoopingErrors", labels, nil),
	}
}

func (s *SwitchCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.PortXmitData
	ch <- s.PortRcvData
	ch <- s.PortXmitPkts
	ch <- s.PortRcvPkts
	ch <- s.PortUnicastXmitPkts
	ch <- s.PortUnicastRcvPkts
	ch <- s.PortMulticastXmitPkts
	ch <- s.PortMulticastRcvPkts
	ch <- s.SymbolErrorCounter
	ch <- s.LinkErrorRecoveryCounter
	ch <- s.LinkDownedCounter
	ch <- s.PortRcvErrors
	ch <- s.PortRcvRemotePhysicalErrors
	ch <- s.PortRcvSwitchRelayErrors
	ch <- s.PortXmitDiscards
	ch <- s.PortXmitConstraintErrors
	ch <- s.PortRcvConstraintErrors
	ch <- s.LocalLinkIntegrityErrors
	ch <- s.ExcessiveBufferOverrunErrors
	ch <- s.VL15Dropped
	ch <- s.PortXmitWait
	ch <- s.QP1Dropped
	ch <- s.PortLocalPhysicalErrors
	ch <- s.PortMalformedPktErrors
	ch <- s.PortBufferOverrunErrors
	ch <- s.PortDLIDMappingErrors
	ch <- s.PortVLMappingErrors
	ch <- s.PortLoopingErrors
}

func (s *SwitchCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()
	counters, errors, timeouts := s.collect()
	for _, c := range counters {
		if !math.IsNaN(c.PortXmitData) {
			ch <- prometheus.MustNewConstMetric(s.PortXmitData, prometheus.CounterValue, c.PortXmitData, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortRcvData) {
			ch <- prometheus.MustNewConstMetric(s.PortRcvData, prometheus.CounterValue, c.PortRcvData, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortXmitPkts) {
			ch <- prometheus.MustNewConstMetric(s.PortXmitPkts, prometheus.CounterValue, c.PortXmitPkts, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortRcvPkts) {
			ch <- prometheus.MustNewConstMetric(s.PortRcvPkts, prometheus.CounterValue, c.PortRcvPkts, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortUnicastXmitPkts) {
			ch <- prometheus.MustNewConstMetric(s.PortUnicastXmitPkts, prometheus.CounterValue, c.PortUnicastXmitPkts, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortUnicastRcvPkts) {
			ch <- prometheus.MustNewConstMetric(s.PortUnicastRcvPkts, prometheus.CounterValue, c.PortUnicastRcvPkts, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortMulticastXmitPkts) {
			ch <- prometheus.MustNewConstMetric(s.PortMulticastXmitPkts, prometheus.CounterValue, c.PortMulticastXmitPkts, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortMulticastRcvPkts) {
			ch <- prometheus.MustNewConstMetric(s.PortMulticastRcvPkts, prometheus.CounterValue, c.PortMulticastRcvPkts, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.SymbolErrorCounter) {
			ch <- prometheus.MustNewConstMetric(s.SymbolErrorCounter, prometheus.CounterValue, c.SymbolErrorCounter, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.LinkErrorRecoveryCounter) {
			ch <- prometheus.MustNewConstMetric(s.LinkErrorRecoveryCounter, prometheus.CounterValue, c.LinkErrorRecoveryCounter, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.LinkDownedCounter) {
			ch <- prometheus.MustNewConstMetric(s.LinkDownedCounter, prometheus.CounterValue, c.LinkDownedCounter, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortRcvErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortRcvErrors, prometheus.CounterValue, c.PortRcvErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortRcvRemotePhysicalErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortRcvRemotePhysicalErrors, prometheus.CounterValue, c.PortRcvRemotePhysicalErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortRcvSwitchRelayErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortRcvSwitchRelayErrors, prometheus.CounterValue, c.PortRcvSwitchRelayErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortXmitDiscards) {
			ch <- prometheus.MustNewConstMetric(s.PortXmitDiscards, prometheus.CounterValue, c.PortXmitDiscards, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortXmitConstraintErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortXmitConstraintErrors, prometheus.CounterValue, c.PortXmitConstraintErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortRcvConstraintErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortRcvConstraintErrors, prometheus.CounterValue, c.PortRcvConstraintErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.LocalLinkIntegrityErrors) {
			ch <- prometheus.MustNewConstMetric(s.LocalLinkIntegrityErrors, prometheus.CounterValue, c.LocalLinkIntegrityErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.ExcessiveBufferOverrunErrors) {
			ch <- prometheus.MustNewConstMetric(s.ExcessiveBufferOverrunErrors, prometheus.CounterValue, c.ExcessiveBufferOverrunErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.VL15Dropped) {
			ch <- prometheus.MustNewConstMetric(s.VL15Dropped, prometheus.CounterValue, c.VL15Dropped, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortXmitWait) {
			ch <- prometheus.MustNewConstMetric(s.PortXmitWait, prometheus.CounterValue, c.PortXmitWait, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.QP1Dropped) {
			ch <- prometheus.MustNewConstMetric(s.QP1Dropped, prometheus.CounterValue, c.QP1Dropped, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortLocalPhysicalErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortLocalPhysicalErrors, prometheus.CounterValue, c.PortLocalPhysicalErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortMalformedPktErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortMalformedPktErrors, prometheus.CounterValue, c.PortMalformedPktErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortBufferOverrunErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortBufferOverrunErrors, prometheus.CounterValue, c.PortBufferOverrunErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortDLIDMappingErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortDLIDMappingErrors, prometheus.CounterValue, c.PortDLIDMappingErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortVLMappingErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortVLMappingErrors, prometheus.CounterValue, c.PortVLMappingErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
		if !math.IsNaN(c.PortLoopingErrors) {
			ch <- prometheus.MustNewConstMetric(s.PortLoopingErrors, prometheus.CounterValue, c.PortLoopingErrors, c.device.GUID, c.PortSelect, c.device.Name)
		}
	}
	ch <- prometheus.MustNewConstMetric(collectErrors, prometheus.GaugeValue, errors, "switch")
	ch <- prometheus.MustNewConstMetric(collecTimeouts, prometheus.GaugeValue, timeouts, "switch")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "switch")
}

func (s *SwitchCollector) collect() ([]PerfQueryCounters, float64, float64) {
	var counters []PerfQueryCounters
	var countersLock sync.Mutex
	var errors, timeouts float64
	limit := make(chan int, *maxConcurrent)
	wg := &sync.WaitGroup{}
	for _, device := range *s.switches {
		limit <- 1
		wg.Add(1)
		go func(device InfinibandDevice) {
			defer wg.Done()
			ctxExtended, cancelExtended := context.WithTimeout(context.Background(), *perfqueryTimeout)
			defer cancelExtended()
			ports := getDevicePorts(device.Uplinks)
			perfqueryPorts := strings.Join(ports, ",")
			extendedOut, err := PerfqueryExec(device.GUID, perfqueryPorts, []string{"-l", "-x"}, ctxExtended)
			if err == context.DeadlineExceeded {
				level.Error(s.logger).Log("msg", "Timeout collecting extended perfquery counters", "guid", device.GUID)
				timeouts++
			} else if err != nil {
				level.Error(s.logger).Log("msg", "Error collecting extended perfquery counters", "guid", device.GUID)
				errors++
			}
			if err != nil {
				<-limit
				return
			}
			deviceCounters, errs := perfqueryParse(device, extendedOut, s.logger)
			errors = errors + errs
			level.Debug(s.logger).Log("msg", "Adding parsed counters", "count", len(deviceCounters), "guid", device.GUID, "name", device.Name)
			countersLock.Lock()
			counters = append(counters, deviceCounters...)
			countersLock.Unlock()
			if *switchCollectRcvErr {
				for _, deviceCounter := range deviceCounters {
					ctxRcvErr, cancelRcvErr := context.WithTimeout(context.Background(), *perfqueryTimeout)
					defer cancelRcvErr()
					rcvErrOut, err := PerfqueryExec(device.GUID, deviceCounter.PortSelect, []string{"-E"}, ctxRcvErr)
					if err == context.DeadlineExceeded {
						level.Error(s.logger).Log("msg", "Timeout collecting rcvErr perfquery counters", "guid", device.GUID)
						timeouts++
						continue
					} else if err != nil {
						level.Error(s.logger).Log("msg", "Error collecting rcvErr perfquery counters", "guid", device.GUID)
						errors++
						continue
					}
					rcvErrCounters, errs := perfqueryParse(device, rcvErrOut, s.logger)
					errors = errors + errs
					countersLock.Lock()
					counters = append(counters, rcvErrCounters...)
					countersLock.Unlock()
				}
			}
			<-limit
		}(device)
	}
	wg.Wait()
	close(limit)
	return counters, errors, timeouts
}
