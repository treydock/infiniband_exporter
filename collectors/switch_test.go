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
	"strings"
	"testing"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSwitchCollector(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, false, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 0
		# HELP infiniband_switch_info Infiniband switch information
		# TYPE infiniband_switch_info gauge
		infiniband_switch_info{guid="0x506b4b03005c2740",lid="2052",switch="ib-i4l1s01"} 1
		infiniband_switch_info{guid="0x7cfe9003009ce5b0",lid="1719",switch="ib-i1l1s01"} 1
		# HELP infiniband_switch_port_excessive_buffer_overrun_errors_total Infiniband switch port ExcessiveBufferOverrunErrors
		# TYPE infiniband_switch_port_excessive_buffer_overrun_errors_total counter
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_link_downed_total Infiniband switch port LinkDownedCounter
		# TYPE infiniband_switch_port_link_downed_total counter
		infiniband_switch_port_link_downed_total{guid="0x506b4b03005c2740",port="1"} 1
		infiniband_switch_port_link_downed_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_link_downed_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_link_error_recovery_total Infiniband switch port LinkErrorRecoveryCounter
		# TYPE infiniband_switch_port_link_error_recovery_total counter
		infiniband_switch_port_link_error_recovery_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_link_error_recovery_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_link_error_recovery_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_local_link_integrity_errors_total Infiniband switch port LocalLinkIntegrityErrors
		# TYPE infiniband_switch_port_local_link_integrity_errors_total counter
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_multicast_receive_packets_total Infiniband switch port PortMulticastRcvPkts
		# TYPE infiniband_switch_port_multicast_receive_packets_total counter
		infiniband_switch_port_multicast_receive_packets_total{guid="0x506b4b03005c2740",port="1"} 6694940
		infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 5584846741
		infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_multicast_transmit_packets_total Infiniband switch port PortMulticastXmitPkts
		# TYPE infiniband_switch_port_multicast_transmit_packets_total counter
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1"} 5623645694
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 25038914
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_qp1_dropped_total Infiniband switch port QP1Dropped
		# TYPE infiniband_switch_port_qp1_dropped_total counter
		infiniband_switch_port_qp1_dropped_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_qp1_dropped_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_qp1_dropped_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_rate_bytes_per_second Infiniband switch port rate
		# TYPE infiniband_switch_port_rate_bytes_per_second gauge
		infiniband_switch_port_rate_bytes_per_second{guid="0x506b4b03005c2740",port="35"} 1.25e+10
		infiniband_switch_port_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="10"} 1.25e+10
		infiniband_switch_port_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="11"} 1.25e+10
		infiniband_switch_port_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="1"} 1.25e+10
		# HELP infiniband_switch_port_raw_rate_bytes_per_second Infiniband switch port raw rate
		# TYPE infiniband_switch_port_raw_rate_bytes_per_second gauge
		infiniband_switch_port_raw_rate_bytes_per_second{guid="0x506b4b03005c2740",port="35"} 1.2890625e+10
		infiniband_switch_port_raw_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="10"} 1.2890625e+10
		infiniband_switch_port_raw_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="11"} 1.2890625e+10
		infiniband_switch_port_raw_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="1"} 1.2890625e+10
		# HELP infiniband_switch_port_receive_constraint_errors_total Infiniband switch port PortRcvConstraintErrors
		# TYPE infiniband_switch_port_receive_constraint_errors_total counter
		infiniband_switch_port_receive_constraint_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_receive_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_receive_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_receive_data_bytes_total Infiniband switch port PortRcvData
		# TYPE infiniband_switch_port_receive_data_bytes_total counter
		infiniband_switch_port_receive_data_bytes_total{guid="0x506b4b03005c2740",port="1"} 178762341961629
		infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1"} 12279028775751
		infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2"} 39078804993378
		# HELP infiniband_switch_port_receive_errors_total Infiniband switch port PortRcvErrors
		# TYPE infiniband_switch_port_receive_errors_total counter
		infiniband_switch_port_receive_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_receive_packets_total Infiniband switch port PortRcvPkts
		# TYPE infiniband_switch_port_receive_packets_total counter
		infiniband_switch_port_receive_packets_total{guid="0x506b4b03005c2740",port="1"} 387654829341
		infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 32262508468
		infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 93660802641
		# HELP infiniband_switch_port_receive_remote_physical_errors_total Infiniband switch port PortRcvRemotePhysicalErrors
		# TYPE infiniband_switch_port_receive_remote_physical_errors_total counter
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_receive_switch_relay_errors_total Infiniband switch port PortRcvSwitchRelayErrors
		# TYPE infiniband_switch_port_receive_switch_relay_errors_total counter
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x506b4b03005c2740",port="1"} 7
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_symbol_error_total Infiniband switch port SymbolErrorCounter
		# TYPE infiniband_switch_port_symbol_error_total counter
		infiniband_switch_port_symbol_error_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_symbol_error_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_symbol_error_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_transmit_constraint_errors_total Infiniband switch port PortXmitConstraintErrors
		# TYPE infiniband_switch_port_transmit_constraint_errors_total counter
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_transmit_data_bytes_total Infiniband switch port PortXmitData
		# TYPE infiniband_switch_port_transmit_data_bytes_total counter
		infiniband_switch_port_transmit_data_bytes_total{guid="0x506b4b03005c2740",port="1"} 178791657177235
		infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1"} 36298026860928
		infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2"} 26006570014026
		# HELP infiniband_switch_port_transmit_discards_total Infiniband switch port PortXmitDiscards
		# TYPE infiniband_switch_port_transmit_discards_total counter
		infiniband_switch_port_transmit_discards_total{guid="0x506b4b03005c2740",port="1"} 20046
		infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_transmit_packets_total Infiniband switch port PortXmitPkts
		# TYPE infiniband_switch_port_transmit_packets_total counter
		infiniband_switch_port_transmit_packets_total{guid="0x506b4b03005c2740",port="1"} 393094651266
		infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 101733204203
		infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 122978948297
		# HELP infiniband_switch_port_transmit_wait_total Infiniband switch port PortXmitWait
		# TYPE infiniband_switch_port_transmit_wait_total counter
		infiniband_switch_port_transmit_wait_total{guid="0x506b4b03005c2740",port="1"} 41864608
		infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="1"} 22730501
		infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="2"} 36510964
		# HELP infiniband_switch_port_unicast_receive_packets_total Infiniband switch port PortUnicastRcvPkts
		# TYPE infiniband_switch_port_unicast_receive_packets_total counter
		infiniband_switch_port_unicast_receive_packets_total{guid="0x506b4b03005c2740",port="1"} 387648134400
		infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 26677661727
		infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 93660802641
		# HELP infiniband_switch_port_unicast_transmit_packets_total Infiniband switch port PortUnicastXmitPkts
		# TYPE infiniband_switch_port_unicast_transmit_packets_total counter
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1"} 387471005571
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 101708165289
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 122978948297
		# HELP infiniband_switch_port_vl15_dropped_total Infiniband switch port VL15Dropped
		# TYPE infiniband_switch_port_vl15_dropped_total counter
		infiniband_switch_port_vl15_dropped_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_uplink_info Infiniband switch uplink information
		# TYPE infiniband_switch_uplink_info gauge
		infiniband_switch_uplink_info{guid="0x506b4b03005c2740",port="35",switch="ib-i4l1s01",uplink="p0001 HCA-1",uplink_guid="0x506b4b0300cc02a6",uplink_lid="1432",uplink_port="1",uplink_type="CA"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01",uplink="ib-i1l2s01",uplink_guid="0x7cfe900300b07320",uplink_lid="1516",uplink_port="1",uplink_type="SW"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="10",switch="ib-i1l1s01",uplink="o0001 HCA-1",uplink_guid="0x7cfe9003003b4bde",uplink_lid="134",uplink_port="1",uplink_type="CA"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="11",switch="ib-i1l1s01",uplink="o0002 HCA-1",uplink_guid="0x7cfe9003003b4b96",uplink_lid="133",uplink_port="1",uplink_type="CA"} 1
	`
	collector := NewSwitchCollector(&switchDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 89 {
		t.Errorf("Unexpected collection count %d, expected 89", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_switch_port_multicast_receive_packets_total", "infiniband_switch_port_multicast_transmit_packets_total",
		"infiniband_switch_port_qp1_dropped_total", "infiniband_switch_port_receive_constraint_errors_total",
		"infiniband_switch_port_receive_data_bytes_total", "infiniband_switch_port_receive_errors_total",
		"infiniband_switch_port_receive_packets_total", "infiniband_switch_port_receive_remote_physical_errors_total",
		"infiniband_switch_port_receive_switch_relay_errors_total", "infiniband_switch_port_symbol_error_total",
		"infiniband_switch_port_transmit_constraint_errors_total", "infiniband_switch_port_transmit_data_bytes_total",
		"infiniband_switch_port_transmit_discards_total", "infiniband_switch_port_transmit_packets_total",
		"infiniband_switch_port_transmit_wait_total", "infiniband_switch_port_unicast_receive_packets_total",
		"infiniband_switch_port_unicast_transmit_packets_total", "infiniband_switch_port_vl15_dropped_total",
		"infiniband_switch_port_buffer_overrun_errors_total",
		"infiniband_switch_info", "infiniband_switch_port_rate_bytes_per_second", "infiniband_switch_port_raw_rate_bytes_per_second", "infiniband_switch_uplink_info",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorFull(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{"--collector.switch.rcv-err-details"}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, false, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 0
		# HELP infiniband_switch_info Infiniband switch information
		# TYPE infiniband_switch_info gauge
		infiniband_switch_info{guid="0x506b4b03005c2740",lid="2052",switch="ib-i4l1s01"} 1
		infiniband_switch_info{guid="0x7cfe9003009ce5b0",lid="1719",switch="ib-i1l1s01"} 1
		# HELP infiniband_switch_port_buffer_overrun_errors_total Infiniband switch port PortBufferOverrunErrors
		# TYPE infiniband_switch_port_buffer_overrun_errors_total counter
		infiniband_switch_port_buffer_overrun_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_dli_mapping_errors_total Infiniband switch port PortDLIDMappingErrors
		# TYPE infiniband_switch_port_dli_mapping_errors_total counter
		infiniband_switch_port_dli_mapping_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_dli_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_dli_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_excessive_buffer_overrun_errors_total Infiniband switch port ExcessiveBufferOverrunErrors
		# TYPE infiniband_switch_port_excessive_buffer_overrun_errors_total counter
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_link_downed_total Infiniband switch port LinkDownedCounter
		# TYPE infiniband_switch_port_link_downed_total counter
		infiniband_switch_port_link_downed_total{guid="0x506b4b03005c2740",port="1"} 1
		infiniband_switch_port_link_downed_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_link_downed_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_link_error_recovery_total Infiniband switch port LinkErrorRecoveryCounter
		# TYPE infiniband_switch_port_link_error_recovery_total counter
		infiniband_switch_port_link_error_recovery_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_link_error_recovery_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_link_error_recovery_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_local_link_integrity_errors_total Infiniband switch port LocalLinkIntegrityErrors
		# TYPE infiniband_switch_port_local_link_integrity_errors_total counter
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_local_physical_errors_total Infiniband switch port PortLocalPhysicalErrors
		# TYPE infiniband_switch_port_local_physical_errors_total counter
		infiniband_switch_port_local_physical_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_local_physical_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_local_physical_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_looping_errors_total Infiniband switch port PortLoopingErrors
		# TYPE infiniband_switch_port_looping_errors_total counter
		infiniband_switch_port_looping_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_looping_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_looping_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_malformed_packet_errors_total Infiniband switch port PortMalformedPktErrors
		# TYPE infiniband_switch_port_malformed_packet_errors_total counter
		infiniband_switch_port_malformed_packet_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_malformed_packet_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_malformed_packet_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_multicast_receive_packets_total Infiniband switch port PortMulticastRcvPkts
		# TYPE infiniband_switch_port_multicast_receive_packets_total counter
		infiniband_switch_port_multicast_receive_packets_total{guid="0x506b4b03005c2740",port="1"} 6694940
		infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 5584846741
		infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_multicast_transmit_packets_total Infiniband switch port PortMulticastXmitPkts
		# TYPE infiniband_switch_port_multicast_transmit_packets_total counter
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1"} 5623645694
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 25038914
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_qp1_dropped_total Infiniband switch port QP1Dropped
		# TYPE infiniband_switch_port_qp1_dropped_total counter
		infiniband_switch_port_qp1_dropped_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_qp1_dropped_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_qp1_dropped_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_rate_bytes_per_second Infiniband switch port rate
		# TYPE infiniband_switch_port_rate_bytes_per_second gauge
		infiniband_switch_port_rate_bytes_per_second{guid="0x506b4b03005c2740",port="35"} 1.25e+10
		infiniband_switch_port_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="10"} 1.25e+10
		infiniband_switch_port_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="11"} 1.25e+10
		infiniband_switch_port_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="1"} 1.25e+10
		# HELP infiniband_switch_port_raw_rate_bytes_per_second Infiniband switch port raw rate
		# TYPE infiniband_switch_port_raw_rate_bytes_per_second gauge
		infiniband_switch_port_raw_rate_bytes_per_second{guid="0x506b4b03005c2740",port="35"} 1.2890625e+10
		infiniband_switch_port_raw_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="10"} 1.2890625e+10
		infiniband_switch_port_raw_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="11"} 1.2890625e+10
		infiniband_switch_port_raw_rate_bytes_per_second{guid="0x7cfe9003009ce5b0",port="1"} 1.2890625e+10
		# HELP infiniband_switch_port_receive_constraint_errors_total Infiniband switch port PortRcvConstraintErrors
		# TYPE infiniband_switch_port_receive_constraint_errors_total counter
		infiniband_switch_port_receive_constraint_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_receive_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_receive_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_receive_data_bytes_total Infiniband switch port PortRcvData
		# TYPE infiniband_switch_port_receive_data_bytes_total counter
		infiniband_switch_port_receive_data_bytes_total{guid="0x506b4b03005c2740",port="1"} 178762341961629
		infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1"} 12279028775751
		infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2"} 39078804993378
		# HELP infiniband_switch_port_receive_errors_total Infiniband switch port PortRcvErrors
		# TYPE infiniband_switch_port_receive_errors_total counter
		infiniband_switch_port_receive_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_receive_packets_total Infiniband switch port PortRcvPkts
		# TYPE infiniband_switch_port_receive_packets_total counter
		infiniband_switch_port_receive_packets_total{guid="0x506b4b03005c2740",port="1"} 387654829341
		infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 32262508468
		infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 93660802641
		# HELP infiniband_switch_port_receive_remote_physical_errors_total Infiniband switch port PortRcvRemotePhysicalErrors
		# TYPE infiniband_switch_port_receive_remote_physical_errors_total counter
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_receive_switch_relay_errors_total Infiniband switch port PortRcvSwitchRelayErrors
		# TYPE infiniband_switch_port_receive_switch_relay_errors_total counter
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x506b4b03005c2740",port="1"} 7
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_symbol_error_total Infiniband switch port SymbolErrorCounter
		# TYPE infiniband_switch_port_symbol_error_total counter
		infiniband_switch_port_symbol_error_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_symbol_error_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_symbol_error_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_transmit_constraint_errors_total Infiniband switch port PortXmitConstraintErrors
		# TYPE infiniband_switch_port_transmit_constraint_errors_total counter
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_transmit_data_bytes_total Infiniband switch port PortXmitData
		# TYPE infiniband_switch_port_transmit_data_bytes_total counter
		infiniband_switch_port_transmit_data_bytes_total{guid="0x506b4b03005c2740",port="1"} 178791657177235
		infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1"} 36298026860928
		infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2"} 26006570014026
		# HELP infiniband_switch_port_transmit_discards_total Infiniband switch port PortXmitDiscards
		# TYPE infiniband_switch_port_transmit_discards_total counter
		infiniband_switch_port_transmit_discards_total{guid="0x506b4b03005c2740",port="1"} 20046
		infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_transmit_packets_total Infiniband switch port PortXmitPkts
		# TYPE infiniband_switch_port_transmit_packets_total counter
		infiniband_switch_port_transmit_packets_total{guid="0x506b4b03005c2740",port="1"} 393094651266
		infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 101733204203
		infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 122978948297
		# HELP infiniband_switch_port_transmit_wait_total Infiniband switch port PortXmitWait
		# TYPE infiniband_switch_port_transmit_wait_total counter
		infiniband_switch_port_transmit_wait_total{guid="0x506b4b03005c2740",port="1"} 41864608
		infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="1"} 22730501
		infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="2"} 36510964
		# HELP infiniband_switch_port_unicast_receive_packets_total Infiniband switch port PortUnicastRcvPkts
		# TYPE infiniband_switch_port_unicast_receive_packets_total counter
		infiniband_switch_port_unicast_receive_packets_total{guid="0x506b4b03005c2740",port="1"} 387648134400
		infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 26677661727
		infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 93660802641
		# HELP infiniband_switch_port_unicast_transmit_packets_total Infiniband switch port PortUnicastXmitPkts
		# TYPE infiniband_switch_port_unicast_transmit_packets_total counter
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1"} 387471005571
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 101708165289
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 122978948297
		# HELP infiniband_switch_port_vl_mapping_errors_total Infiniband switch port PortVLMappingErrors
		# TYPE infiniband_switch_port_vl_mapping_errors_total counter
		infiniband_switch_port_vl_mapping_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_vl_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_vl_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_vl15_dropped_total Infiniband switch port VL15Dropped
		# TYPE infiniband_switch_port_vl15_dropped_total counter
		infiniband_switch_port_vl15_dropped_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_uplink_info Infiniband switch uplink information
		# TYPE infiniband_switch_uplink_info gauge
		infiniband_switch_uplink_info{guid="0x506b4b03005c2740",port="35",switch="ib-i4l1s01",uplink="p0001 HCA-1",uplink_guid="0x506b4b0300cc02a6",uplink_lid="1432",uplink_port="1",uplink_type="CA"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01",uplink="ib-i1l2s01",uplink_guid="0x7cfe900300b07320",uplink_lid="1516",uplink_port="1",uplink_type="SW"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="10",switch="ib-i1l1s01",uplink="o0001 HCA-1",uplink_guid="0x7cfe9003003b4bde",uplink_lid="134",uplink_port="1",uplink_type="CA"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="11",switch="ib-i1l1s01",uplink="o0002 HCA-1",uplink_guid="0x7cfe9003003b4b96",uplink_lid="133",uplink_port="1",uplink_type="CA"} 1
	`
	collector := NewSwitchCollector(&switchDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 113 {
		t.Errorf("Unexpected collection count %d, expected 113", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_switch_port_multicast_receive_packets_total", "infiniband_switch_port_multicast_transmit_packets_total",
		"infiniband_switch_port_qp1_dropped_total", "infiniband_switch_port_receive_constraint_errors_total",
		"infiniband_switch_port_receive_data_bytes_total", "infiniband_switch_port_receive_errors_total",
		"infiniband_switch_port_receive_packets_total", "infiniband_switch_port_receive_remote_physical_errors_total",
		"infiniband_switch_port_receive_switch_relay_errors_total", "infiniband_switch_port_symbol_error_total",
		"infiniband_switch_port_transmit_constraint_errors_total", "infiniband_switch_port_transmit_data_bytes_total",
		"infiniband_switch_port_transmit_discards_total", "infiniband_switch_port_transmit_packets_total",
		"infiniband_switch_port_transmit_wait_total", "infiniband_switch_port_unicast_receive_packets_total",
		"infiniband_switch_port_unicast_transmit_packets_total", "infiniband_switch_port_vl15_dropped_total",
		"infiniband_switch_port_buffer_overrun_errors_total", "infiniband_switch_port_dli_mapping_errors_total",
		"infiniband_switch_port_local_physical_errors_total", "infiniband_switch_port_looping_errors_total",
		"infiniband_switch_port_malformed_packet_errors_total", "infiniband_switch_port_vl_mapping_errors_total",
		"infiniband_switch_info", "infiniband_switch_port_rate_bytes_per_second", "infiniband_switch_port_raw_rate_bytes_per_second", "infiniband_switch_uplink_info",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorNoBase(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{"--no-collector.switch.base-metrics", "--collector.switch.rcv-err-details"}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, false, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 0
		# HELP infiniband_switch_port_buffer_overrun_errors_total Infiniband switch port PortBufferOverrunErrors
		# TYPE infiniband_switch_port_buffer_overrun_errors_total counter
		infiniband_switch_port_buffer_overrun_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_dli_mapping_errors_total Infiniband switch port PortDLIDMappingErrors
		# TYPE infiniband_switch_port_dli_mapping_errors_total counter
		infiniband_switch_port_dli_mapping_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_dli_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_dli_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_local_physical_errors_total Infiniband switch port PortLocalPhysicalErrors
		# TYPE infiniband_switch_port_local_physical_errors_total counter
		infiniband_switch_port_local_physical_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_local_physical_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_local_physical_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_looping_errors_total Infiniband switch port PortLoopingErrors
		# TYPE infiniband_switch_port_looping_errors_total counter
		infiniband_switch_port_looping_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_looping_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_looping_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_malformed_packet_errors_total Infiniband switch port PortMalformedPktErrors
		# TYPE infiniband_switch_port_malformed_packet_errors_total counter
		infiniband_switch_port_malformed_packet_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_malformed_packet_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_malformed_packet_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
		# HELP infiniband_switch_port_vl_mapping_errors_total Infiniband switch port PortVLMappingErrors
		# TYPE infiniband_switch_port_vl_mapping_errors_total counter
		infiniband_switch_port_vl_mapping_errors_total{guid="0x506b4b03005c2740",port="1"} 0
		infiniband_switch_port_vl_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
		infiniband_switch_port_vl_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
	`
	collector := NewSwitchCollector(&switchDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 27 {
		t.Errorf("Unexpected collection count %d, expected 27", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_switch_port_multicast_receive_packets_total", "infiniband_switch_port_multicast_transmit_packets_total",
		"infiniband_switch_port_qp1_dropped_total", "infiniband_switch_port_receive_constraint_errors_total",
		"infiniband_switch_port_receive_data_bytes_total", "infiniband_switch_port_receive_errors_total",
		"infiniband_switch_port_receive_packets_total", "infiniband_switch_port_receive_remote_physical_errors_total",
		"infiniband_switch_port_receive_switch_relay_errors_total", "infiniband_switch_port_symbol_error_total",
		"infiniband_switch_port_transmit_constraint_errors_total", "infiniband_switch_port_transmit_data_bytes_total",
		"infiniband_switch_port_transmit_discards_total", "infiniband_switch_port_transmit_packets_total",
		"infiniband_switch_port_transmit_wait_total", "infiniband_switch_port_unicast_receive_packets_total",
		"infiniband_switch_port_unicast_transmit_packets_total", "infiniband_switch_port_vl15_dropped_total",
		"infiniband_switch_port_buffer_overrun_errors_total", "infiniband_switch_port_dli_mapping_errors_total",
		"infiniband_switch_port_local_physical_errors_total", "infiniband_switch_port_looping_errors_total",
		"infiniband_switch_port_malformed_packet_errors_total", "infiniband_switch_port_vl_mapping_errors_total",
		"infiniband_switch_info", "infiniband_switch_port_rate_bytes_per_second", "infiniband_switch_raw_port_rate_bytes_per_second", "infiniband_switch_uplink_info",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorError(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, true, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 2
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 0
	`
	collector := NewSwitchCollector(&switchDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 23 {
		t.Errorf("Unexpected collection count %d, expected 23", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorErrorRunonce(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, true, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch-runonce"} 2
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch-runonce"} 0
	`
	collector := NewSwitchCollector(&switchDevices, true, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 24 {
		t.Errorf("Unexpected collection count %d, expected 24", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorTimeout(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, false, true)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 2
	`
	collector := NewSwitchCollector(&switchDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 23 {
		t.Errorf("Unexpected collection count %d, expected 23", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
