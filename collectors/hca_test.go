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

var (
	hcaDevices = []InfinibandDevice{
		{Type: "CA", LID: "133", GUID: "0x7cfe9003003b4b96", Rate: (25 * 4 * 125000000), RawRate: 1.2890625e+10, Name: "o0002 HCA-1",
			Uplinks: map[string]InfinibandUplink{
				"1": {Type: "SW", LID: "1719", PortNumber: "11", GUID: "0x7cfe9003009ce5b0", Name: "ib-i1l1s01"},
			},
		},
		{Type: "CA", LID: "134", GUID: "0x7cfe9003003b4bde", Rate: (25 * 4 * 125000000), RawRate: 1.2890625e+10, Name: "o0001 HCA-1",
			Uplinks: map[string]InfinibandUplink{
				"1": {Type: "SW", LID: "1719", PortNumber: "10", GUID: "0x7cfe9003009ce5b0", Name: "ib-i1l1s01"},
			},
		},
	}
)

func TestHCACollector(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, false, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="hca"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="hca"} 0
		# HELP infiniband_hca_info Infiniband HCA information
		# TYPE infiniband_hca_info gauge
		infiniband_hca_info{guid="0x7cfe9003003b4b96",hca="o0002 HCA-1",lid="133"} 1
		infiniband_hca_info{guid="0x7cfe9003003b4bde",hca="o0001 HCA-1",lid="134"} 1
		# HELP infiniband_hca_port_excessive_buffer_overrun_errors_total Infiniband HCA port ExcessiveBufferOverrunErrors
		# TYPE infiniband_hca_port_excessive_buffer_overrun_errors_total counter
		infiniband_hca_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_link_downed_total Infiniband HCA port LinkDownedCounter
		# TYPE infiniband_hca_port_link_downed_total counter
		infiniband_hca_port_link_downed_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_link_downed_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_link_error_recovery_total Infiniband HCA port LinkErrorRecoveryCounter
		# TYPE infiniband_hca_port_link_error_recovery_total counter
		infiniband_hca_port_link_error_recovery_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_link_error_recovery_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_local_link_integrity_errors_total Infiniband HCA port LocalLinkIntegrityErrors
		# TYPE infiniband_hca_port_local_link_integrity_errors_total counter
		infiniband_hca_port_local_link_integrity_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_local_link_integrity_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_multicast_receive_packets_total Infiniband HCA port PortMulticastRcvPkts
		# TYPE infiniband_hca_port_multicast_receive_packets_total counter
		infiniband_hca_port_multicast_receive_packets_total{guid="0x7cfe9003003b4b96",port="1"} 3732373137
		infiniband_hca_port_multicast_receive_packets_total{guid="0x7cfe9003003b4bde",port="1"} 3732158589
		# HELP infiniband_hca_port_multicast_transmit_packets_total Infiniband HCA port PortMulticastXmitPkts
		# TYPE infiniband_hca_port_multicast_transmit_packets_total counter
		infiniband_hca_port_multicast_transmit_packets_total{guid="0x7cfe9003003b4b96",port="1"} 544690
		infiniband_hca_port_multicast_transmit_packets_total{guid="0x7cfe9003003b4bde",port="1"} 721488
		# HELP infiniband_hca_port_qp1_dropped_total Infiniband HCA port QP1Dropped
		# TYPE infiniband_hca_port_qp1_dropped_total counter
		infiniband_hca_port_qp1_dropped_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_qp1_dropped_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_receive_constraint_errors_total Infiniband HCA port PortRcvConstraintErrors
		# TYPE infiniband_hca_port_receive_constraint_errors_total counter
		infiniband_hca_port_receive_constraint_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_receive_constraint_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_receive_data_bytes_total Infiniband HCA port PortRcvData
		# TYPE infiniband_hca_port_receive_data_bytes_total counter
		infiniband_hca_port_receive_data_bytes_total{guid="0x7cfe9003003b4b96",port="1"} 37225401952885
		infiniband_hca_port_receive_data_bytes_total{guid="0x7cfe9003003b4bde",port="1"} 9752484588300
		# HELP infiniband_hca_port_receive_errors_total Infiniband HCA port PortRcvErrors
		# TYPE infiniband_hca_port_receive_errors_total counter
		infiniband_hca_port_receive_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_receive_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_receive_packets_total Infiniband HCA port PortRcvPkts
		# TYPE infiniband_hca_port_receive_packets_total counter
		infiniband_hca_port_receive_packets_total{guid="0x7cfe9003003b4b96",port="1"} 100583719365
		infiniband_hca_port_receive_packets_total{guid="0x7cfe9003003b4bde",port="1"} 33038722564
		# HELP infiniband_hca_port_receive_remote_physical_errors_total Infiniband HCA port PortRcvRemotePhysicalErrors
		# TYPE infiniband_hca_port_receive_remote_physical_errors_total counter
		infiniband_hca_port_receive_remote_physical_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_receive_remote_physical_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_receive_switch_relay_errors_total Infiniband HCA port PortRcvSwitchRelayErrors
		# TYPE infiniband_hca_port_receive_switch_relay_errors_total counter
		infiniband_hca_port_receive_switch_relay_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_receive_switch_relay_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_symbol_error_total Infiniband HCA port SymbolErrorCounter
		# TYPE infiniband_hca_port_symbol_error_total counter
		infiniband_hca_port_symbol_error_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_symbol_error_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_transmit_constraint_errors_total Infiniband HCA port PortXmitConstraintErrors
		# TYPE infiniband_hca_port_transmit_constraint_errors_total counter
		infiniband_hca_port_transmit_constraint_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_transmit_constraint_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_transmit_data_bytes_total Infiniband HCA port PortXmitData
		# TYPE infiniband_hca_port_transmit_data_bytes_total counter
		infiniband_hca_port_transmit_data_bytes_total{guid="0x7cfe9003003b4b96",port="1"} 37108676853855
		infiniband_hca_port_transmit_data_bytes_total{guid="0x7cfe9003003b4bde",port="1"} 9049592493976
		# HELP infiniband_hca_port_transmit_discards_total Infiniband HCA port PortXmitDiscards
		# TYPE infiniband_hca_port_transmit_discards_total counter
		infiniband_hca_port_transmit_discards_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_transmit_discards_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_transmit_packets_total Infiniband HCA port PortXmitPkts
		# TYPE infiniband_hca_port_transmit_packets_total counter
		infiniband_hca_port_transmit_packets_total{guid="0x7cfe9003003b4b96",port="1"} 96917117320
		infiniband_hca_port_transmit_packets_total{guid="0x7cfe9003003b4bde",port="1"} 28825338611
		# HELP infiniband_hca_port_transmit_wait_total Infiniband HCA port PortXmitWait
		# TYPE infiniband_hca_port_transmit_wait_total counter
		infiniband_hca_port_transmit_wait_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_transmit_wait_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_unicast_receive_packets_total Infiniband HCA port PortUnicastRcvPkts
		# TYPE infiniband_hca_port_unicast_receive_packets_total counter
		infiniband_hca_port_unicast_receive_packets_total{guid="0x7cfe9003003b4b96",port="1"} 96851346228
		infiniband_hca_port_unicast_receive_packets_total{guid="0x7cfe9003003b4bde",port="1"} 29306563974
		# HELP infiniband_hca_port_unicast_transmit_packets_total Infiniband HCA port PortUnicastXmitPkts
		# TYPE infiniband_hca_port_unicast_transmit_packets_total counter
		infiniband_hca_port_unicast_transmit_packets_total{guid="0x7cfe9003003b4b96",port="1"} 96916572630
		infiniband_hca_port_unicast_transmit_packets_total{guid="0x7cfe9003003b4bde",port="1"} 28824617123
		# HELP infiniband_hca_port_vl15_dropped_total Infiniband HCA port VL15Dropped
		# TYPE infiniband_hca_port_vl15_dropped_total counter
		infiniband_hca_port_vl15_dropped_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_vl15_dropped_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_rate_bytes_per_second Infiniband HCA rate
		# TYPE infiniband_hca_rate_bytes_per_second gauge
		infiniband_hca_rate_bytes_per_second{guid="0x7cfe9003003b4b96"} 1.25e+10
		infiniband_hca_rate_bytes_per_second{guid="0x7cfe9003003b4bde"} 1.25e+10
		# HELP infiniband_hca_raw_rate_bytes_per_second Infiniband HCA raw rate
		# TYPE infiniband_hca_raw_rate_bytes_per_second gauge
		infiniband_hca_raw_rate_bytes_per_second{guid="0x7cfe9003003b4b96"} 1.2890625e+10
		infiniband_hca_raw_rate_bytes_per_second{guid="0x7cfe9003003b4bde"} 1.2890625e+10
		# HELP infiniband_hca_uplink_info Infiniband HCA uplink information
		# TYPE infiniband_hca_uplink_info gauge
		infiniband_hca_uplink_info{guid="0x7cfe9003003b4b96",hca="o0002 HCA-1",port="1",uplink="ib-i1l1s01",uplink_guid="0x7cfe9003009ce5b0",uplink_lid="1719",uplink_port="11",uplink_type="SW"} 1
		infiniband_hca_uplink_info{guid="0x7cfe9003003b4bde",hca="o0001 HCA-1",port="1",uplink="ib-i1l1s01",uplink_guid="0x7cfe9003009ce5b0",uplink_lid="1719",uplink_port="10",uplink_type="SW"} 1
	`
	collector := NewHCACollector(&hcaDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 55 {
		t.Errorf("Unexpected collection count %d, expected 55", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_hca_port_excessive_buffer_overrun_errors_total", "infiniband_hca_port_link_downed_total",
		"infiniband_hca_port_link_error_recovery_total", "infiniband_hca_port_local_link_integrity_errors_total",
		"infiniband_hca_port_multicast_receive_packets_total", "infiniband_hca_port_multicast_transmit_packets_total",
		"infiniband_hca_port_qp1_dropped_total", "infiniband_hca_port_receive_constraint_errors_total",
		"infiniband_hca_port_receive_data_bytes_total", "infiniband_hca_port_receive_errors_total",
		"infiniband_hca_port_receive_packets_total", "infiniband_hca_port_receive_remote_physical_errors_total",
		"infiniband_hca_port_receive_switch_relay_errors_total", "infiniband_hca_port_symbol_error_total",
		"infiniband_hca_port_transmit_constraint_errors_total", "infiniband_hca_port_transmit_data_bytes_total",
		"infiniband_hca_port_transmit_discards_total", "infiniband_hca_port_transmit_packets_total",
		"infiniband_hca_port_transmit_wait_total", "infiniband_hca_port_unicast_receive_packets_total",
		"infiniband_hca_port_unicast_transmit_packets_total", "infiniband_hca_port_vl15_dropped_total",
		"infiniband_hca_port_buffer_overrun_errors_total",
		"infiniband_hca_info", "infiniband_hca_rate_bytes_per_second", "infiniband_hca_raw_rate_bytes_per_second", "infiniband_hca_uplink_info",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestHCACollectorFull(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{"--collector.hca.rcv-err-details"}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, false, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="hca"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="hca"} 0
		# HELP infiniband_hca_info Infiniband HCA information
		# TYPE infiniband_hca_info gauge
		infiniband_hca_info{guid="0x7cfe9003003b4b96",hca="o0002 HCA-1",lid="133"} 1
		infiniband_hca_info{guid="0x7cfe9003003b4bde",hca="o0001 HCA-1",lid="134"} 1
		# HELP infiniband_hca_port_buffer_overrun_errors_total Infiniband HCA port PortBufferOverrunErrors
		# TYPE infiniband_hca_port_buffer_overrun_errors_total counter
		infiniband_hca_port_buffer_overrun_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_buffer_overrun_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_dli_mapping_errors_total Infiniband HCA port PortDLIDMappingErrors
		# TYPE infiniband_hca_port_dli_mapping_errors_total counter
		infiniband_hca_port_dli_mapping_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_dli_mapping_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_excessive_buffer_overrun_errors_total Infiniband HCA port ExcessiveBufferOverrunErrors
		# TYPE infiniband_hca_port_excessive_buffer_overrun_errors_total counter
		infiniband_hca_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_link_downed_total Infiniband HCA port LinkDownedCounter
		# TYPE infiniband_hca_port_link_downed_total counter
		infiniband_hca_port_link_downed_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_link_downed_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_link_error_recovery_total Infiniband HCA port LinkErrorRecoveryCounter
		# TYPE infiniband_hca_port_link_error_recovery_total counter
		infiniband_hca_port_link_error_recovery_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_link_error_recovery_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_local_link_integrity_errors_total Infiniband HCA port LocalLinkIntegrityErrors
		# TYPE infiniband_hca_port_local_link_integrity_errors_total counter
		infiniband_hca_port_local_link_integrity_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_local_link_integrity_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_local_physical_errors_total Infiniband HCA port PortLocalPhysicalErrors
		# TYPE infiniband_hca_port_local_physical_errors_total counter
		infiniband_hca_port_local_physical_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_local_physical_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_looping_errors_total Infiniband HCA port PortLoopingErrors
		# TYPE infiniband_hca_port_looping_errors_total counter
		infiniband_hca_port_looping_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_looping_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_malformed_packet_errors_total Infiniband HCA port PortMalformedPktErrors
		# TYPE infiniband_hca_port_malformed_packet_errors_total counter
		infiniband_hca_port_malformed_packet_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_malformed_packet_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_multicast_receive_packets_total Infiniband HCA port PortMulticastRcvPkts
		# TYPE infiniband_hca_port_multicast_receive_packets_total counter
		infiniband_hca_port_multicast_receive_packets_total{guid="0x7cfe9003003b4b96",port="1"} 3732373137
		infiniband_hca_port_multicast_receive_packets_total{guid="0x7cfe9003003b4bde",port="1"} 3732158589
		# HELP infiniband_hca_port_multicast_transmit_packets_total Infiniband HCA port PortMulticastXmitPkts
		# TYPE infiniband_hca_port_multicast_transmit_packets_total counter
		infiniband_hca_port_multicast_transmit_packets_total{guid="0x7cfe9003003b4b96",port="1"} 544690
		infiniband_hca_port_multicast_transmit_packets_total{guid="0x7cfe9003003b4bde",port="1"} 721488
		# HELP infiniband_hca_port_qp1_dropped_total Infiniband HCA port QP1Dropped
		# TYPE infiniband_hca_port_qp1_dropped_total counter
		infiniband_hca_port_qp1_dropped_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_qp1_dropped_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_receive_constraint_errors_total Infiniband HCA port PortRcvConstraintErrors
		# TYPE infiniband_hca_port_receive_constraint_errors_total counter
		infiniband_hca_port_receive_constraint_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_receive_constraint_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_receive_data_bytes_total Infiniband HCA port PortRcvData
		# TYPE infiniband_hca_port_receive_data_bytes_total counter
		infiniband_hca_port_receive_data_bytes_total{guid="0x7cfe9003003b4b96",port="1"} 37225401952885
		infiniband_hca_port_receive_data_bytes_total{guid="0x7cfe9003003b4bde",port="1"} 9752484588300
		# HELP infiniband_hca_port_receive_errors_total Infiniband HCA port PortRcvErrors
		# TYPE infiniband_hca_port_receive_errors_total counter
		infiniband_hca_port_receive_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_receive_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_receive_packets_total Infiniband HCA port PortRcvPkts
		# TYPE infiniband_hca_port_receive_packets_total counter
		infiniband_hca_port_receive_packets_total{guid="0x7cfe9003003b4b96",port="1"} 100583719365
		infiniband_hca_port_receive_packets_total{guid="0x7cfe9003003b4bde",port="1"} 33038722564
		# HELP infiniband_hca_port_receive_remote_physical_errors_total Infiniband HCA port PortRcvRemotePhysicalErrors
		# TYPE infiniband_hca_port_receive_remote_physical_errors_total counter
		infiniband_hca_port_receive_remote_physical_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_receive_remote_physical_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_receive_switch_relay_errors_total Infiniband HCA port PortRcvSwitchRelayErrors
		# TYPE infiniband_hca_port_receive_switch_relay_errors_total counter
		infiniband_hca_port_receive_switch_relay_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_receive_switch_relay_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_symbol_error_total Infiniband HCA port SymbolErrorCounter
		# TYPE infiniband_hca_port_symbol_error_total counter
		infiniband_hca_port_symbol_error_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_symbol_error_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_transmit_constraint_errors_total Infiniband HCA port PortXmitConstraintErrors
		# TYPE infiniband_hca_port_transmit_constraint_errors_total counter
		infiniband_hca_port_transmit_constraint_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_transmit_constraint_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_transmit_data_bytes_total Infiniband HCA port PortXmitData
		# TYPE infiniband_hca_port_transmit_data_bytes_total counter
		infiniband_hca_port_transmit_data_bytes_total{guid="0x7cfe9003003b4b96",port="1"} 37108676853855
		infiniband_hca_port_transmit_data_bytes_total{guid="0x7cfe9003003b4bde",port="1"} 9049592493976
		# HELP infiniband_hca_port_transmit_discards_total Infiniband HCA port PortXmitDiscards
		# TYPE infiniband_hca_port_transmit_discards_total counter
		infiniband_hca_port_transmit_discards_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_transmit_discards_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_transmit_packets_total Infiniband HCA port PortXmitPkts
		# TYPE infiniband_hca_port_transmit_packets_total counter
		infiniband_hca_port_transmit_packets_total{guid="0x7cfe9003003b4b96",port="1"} 96917117320
		infiniband_hca_port_transmit_packets_total{guid="0x7cfe9003003b4bde",port="1"} 28825338611
		# HELP infiniband_hca_port_transmit_wait_total Infiniband HCA port PortXmitWait
		# TYPE infiniband_hca_port_transmit_wait_total counter
		infiniband_hca_port_transmit_wait_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_transmit_wait_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_unicast_receive_packets_total Infiniband HCA port PortUnicastRcvPkts
		# TYPE infiniband_hca_port_unicast_receive_packets_total counter
		infiniband_hca_port_unicast_receive_packets_total{guid="0x7cfe9003003b4b96",port="1"} 96851346228
		infiniband_hca_port_unicast_receive_packets_total{guid="0x7cfe9003003b4bde",port="1"} 29306563974
		# HELP infiniband_hca_port_unicast_transmit_packets_total Infiniband HCA port PortUnicastXmitPkts
		# TYPE infiniband_hca_port_unicast_transmit_packets_total counter
		infiniband_hca_port_unicast_transmit_packets_total{guid="0x7cfe9003003b4b96",port="1"} 96916572630
		infiniband_hca_port_unicast_transmit_packets_total{guid="0x7cfe9003003b4bde",port="1"} 28824617123
		# HELP infiniband_hca_port_vl_mapping_errors_total Infiniband HCA port PortVLMappingErrors
		# TYPE infiniband_hca_port_vl_mapping_errors_total counter
		infiniband_hca_port_vl_mapping_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_vl_mapping_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_port_vl15_dropped_total Infiniband HCA port VL15Dropped
		# TYPE infiniband_hca_port_vl15_dropped_total counter
		infiniband_hca_port_vl15_dropped_total{guid="0x7cfe9003003b4b96",port="1"} 0
		infiniband_hca_port_vl15_dropped_total{guid="0x7cfe9003003b4bde",port="1"} 0
		# HELP infiniband_hca_rate_bytes_per_second Infiniband HCA rate
		# TYPE infiniband_hca_rate_bytes_per_second gauge
		infiniband_hca_rate_bytes_per_second{guid="0x7cfe9003003b4b96"} 1.25e+10
		infiniband_hca_rate_bytes_per_second{guid="0x7cfe9003003b4bde"} 1.25e+10
		# HELP infiniband_hca_raw_rate_bytes_per_second Infiniband HCA raw rate
		# TYPE infiniband_hca_raw_rate_bytes_per_second gauge
		infiniband_hca_raw_rate_bytes_per_second{guid="0x7cfe9003003b4b96"} 1.2890625e+10
		infiniband_hca_raw_rate_bytes_per_second{guid="0x7cfe9003003b4bde"} 1.2890625e+10
		# HELP infiniband_hca_uplink_info Infiniband HCA uplink information
		# TYPE infiniband_hca_uplink_info gauge
		infiniband_hca_uplink_info{guid="0x7cfe9003003b4b96",hca="o0002 HCA-1",port="1",uplink="ib-i1l1s01",uplink_guid="0x7cfe9003009ce5b0",uplink_lid="1719",uplink_port="11",uplink_type="SW"} 1
		infiniband_hca_uplink_info{guid="0x7cfe9003003b4bde",hca="o0001 HCA-1",port="1",uplink="ib-i1l1s01",uplink_guid="0x7cfe9003009ce5b0",uplink_lid="1719",uplink_port="10",uplink_type="SW"} 1
	`
	collector := NewHCACollector(&hcaDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 67 {
		t.Errorf("Unexpected collection count %d, expected 67", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_hca_port_excessive_buffer_overrun_errors_total", "infiniband_hca_port_link_downed_total",
		"infiniband_hca_port_link_error_recovery_total", "infiniband_hca_port_local_link_integrity_errors_total",
		"infiniband_hca_port_multicast_receive_packets_total", "infiniband_hca_port_multicast_transmit_packets_total",
		"infiniband_hca_port_qp1_dropped_total", "infiniband_hca_port_receive_constraint_errors_total",
		"infiniband_hca_port_receive_data_bytes_total", "infiniband_hca_port_receive_errors_total",
		"infiniband_hca_port_receive_packets_total", "infiniband_hca_port_receive_remote_physical_errors_total",
		"infiniband_hca_port_receive_switch_relay_errors_total", "infiniband_hca_port_symbol_error_total",
		"infiniband_hca_port_transmit_constraint_errors_total", "infiniband_hca_port_transmit_data_bytes_total",
		"infiniband_hca_port_transmit_discards_total", "infiniband_hca_port_transmit_packets_total",
		"infiniband_hca_port_transmit_wait_total", "infiniband_hca_port_unicast_receive_packets_total",
		"infiniband_hca_port_unicast_transmit_packets_total", "infiniband_hca_port_vl15_dropped_total",
		"infiniband_hca_port_buffer_overrun_errors_total", "infiniband_hca_port_dli_mapping_errors_total",
		"infiniband_hca_port_local_physical_errors_total", "infiniband_hca_port_looping_errors_total",
		"infiniband_hca_port_malformed_packet_errors_total", "infiniband_hca_port_vl_mapping_errors_total",
		"infiniband_hca_info", "infiniband_hca_rate_bytes_per_second", "infiniband_hca_raw_rate_bytes_per_second", "infiniband_hca_uplink_info",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestHCACollectorError(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, true, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="hca"} 2
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="hca"} 0
	`
	collector := NewHCACollector(&hcaDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 11 {
		t.Errorf("Unexpected collection count %d, expected 11", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_hca_port_excessive_buffer_overrun_errors_total", "infiniband_hca_port_link_downed_total",
		"infiniband_hca_port_link_error_recovery_total", "infiniband_hca_port_local_link_integrity_errors_total",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestHCACollectorErrorRunonce(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, true, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="hca-runonce"} 2
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="hca-runonce"} 0
	`
	collector := NewHCACollector(&hcaDevices, true, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 12 {
		t.Errorf("Unexpected collection count %d, expected 12", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_hca_port_excessive_buffer_overrun_errors_total", "infiniband_hca_port_link_downed_total",
		"infiniband_hca_port_link_error_recovery_total", "infiniband_hca_port_local_link_integrity_errors_total",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestHCACollectorTimeout(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, false, true)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="hca"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="hca"} 2
	`
	collector := NewHCACollector(&hcaDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 11 {
		t.Errorf("Unexpected collection count %d, expected 11", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_hca_port_excessive_buffer_overrun_errors_total", "infiniband_hca_port_link_downed_total",
		"infiniband_hca_port_link_error_recovery_total", "infiniband_hca_port_local_link_integrity_errors_total",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
