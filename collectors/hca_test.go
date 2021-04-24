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
	"fmt"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	hcaDevices = []InfinibandDevice{
		InfinibandDevice{GUID: "0x7cfe9003003b4bde", Name: "o0001"},
		InfinibandDevice{GUID: "0x7cfe9003003b4b96", Name: "o0002"},
	}
	perfqueryOutHCA1 = `# Port extended counters: Lid 134 port 1 (CapMask: 0x5A00 CapMask2: 0x0000000)
PortSelect:......................1
CounterSelect:...................0x0000
PortXmitData:....................9049592493976
PortRcvData:.....................9752484588300
PortXmitPkts:....................28825338611
PortRcvPkts:.....................33038722564
PortUnicastXmitPkts:.............28824617123
PortUnicastRcvPkts:..............29306563974
PortMulticastXmitPkts:...........721488
PortMulticastRcvPkts:............3732158589
CounterSelect2:..................0x00000000
SymbolErrorCounter:..............0
LinkErrorRecoveryCounter:........0
LinkDownedCounter:...............0
PortRcvErrors:...................0
PortRcvRemotePhysicalErrors:.....0
PortRcvSwitchRelayErrors:........0
PortXmitDiscards:................0
PortXmitConstraintErrors:........0
PortRcvConstraintErrors:.........0
LocalLinkIntegrityErrors:........0
ExcessiveBufferOverrunErrors:....0
VL15Dropped:.....................0
PortXmitWait:....................0
QP1Dropped:......................0
`
	perfqueryOutHCA2 = `# Port extended counters: Lid 133 port 1 (CapMask: 0x5A00 CapMask2: 0x0000000)
PortSelect:......................1
CounterSelect:...................0x0000
PortXmitData:....................37108676853855
PortRcvData:.....................37225401952885
PortXmitPkts:....................96917117320
PortRcvPkts:.....................100583719365
PortUnicastXmitPkts:.............96916572630
PortUnicastRcvPkts:..............96851346228
PortMulticastXmitPkts:...........544690
PortMulticastRcvPkts:............3732373137
CounterSelect2:..................0x00000000
SymbolErrorCounter:..............0
LinkErrorRecoveryCounter:........0
LinkDownedCounter:...............0
PortRcvErrors:...................0
PortRcvRemotePhysicalErrors:.....0
PortRcvSwitchRelayErrors:........0
PortXmitDiscards:................0
PortXmitConstraintErrors:........0
PortRcvConstraintErrors:.........0
LocalLinkIntegrityErrors:........0
ExcessiveBufferOverrunErrors:....0
VL15Dropped:.....................0
PortXmitWait:....................0
QP1Dropped:......................0
`
)

func TestHCACollector(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	PerfqueryExec = func(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
		if guid == "0x7cfe9003003b4bde" {
			return perfqueryOutHCA1, nil
		} else if guid == "0x7cfe9003003b4b96" {
			return perfqueryOutHCA2, nil
		} else {
			return "", nil
		}
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="hca"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="hca"} 0
		# HELP infiniband_hca_port_excessive_buffer_overrun_errors_total Infiniband HCA port ExcessiveBufferOverrunErrors
		# TYPE infiniband_hca_port_excessive_buffer_overrun_errors_total counter
		infiniband_hca_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_link_downed_total Infiniband HCA port LinkDownedCounter
		# TYPE infiniband_hca_port_link_downed_total counter
		infiniband_hca_port_link_downed_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_link_downed_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_link_error_recovery_total Infiniband HCA port LinkErrorRecoveryCounter
		# TYPE infiniband_hca_port_link_error_recovery_total counter
		infiniband_hca_port_link_error_recovery_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_link_error_recovery_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_local_link_integrity_errors_total Infiniband HCA port LocalLinkIntegrityErrors
		# TYPE infiniband_hca_port_local_link_integrity_errors_total counter
		infiniband_hca_port_local_link_integrity_errors_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_local_link_integrity_errors_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_multicast_receive_packets_total Infiniband HCA port PortMulticastRcvPkts
		# TYPE infiniband_hca_port_multicast_receive_packets_total counter
		infiniband_hca_port_multicast_receive_packets_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 3732373137
		infiniband_hca_port_multicast_receive_packets_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 3732158589
		# HELP infiniband_hca_port_multicast_transmit_packets_total Infiniband HCA port PortMulticastXmitPkts
		# TYPE infiniband_hca_port_multicast_transmit_packets_total counter
		infiniband_hca_port_multicast_transmit_packets_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 544690
		infiniband_hca_port_multicast_transmit_packets_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 721488
		# HELP infiniband_hca_port_qp1_dropped_total Infiniband HCA port QP1Dropped
		# TYPE infiniband_hca_port_qp1_dropped_total counter
		infiniband_hca_port_qp1_dropped_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_qp1_dropped_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_receive_constraint_errors_total Infiniband HCA port PortRcvConstraintErrors
		# TYPE infiniband_hca_port_receive_constraint_errors_total counter
		infiniband_hca_port_receive_constraint_errors_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_receive_constraint_errors_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_receive_data_bytes_total Infiniband HCA port PortRcvData
		# TYPE infiniband_hca_port_receive_data_bytes_total counter
		infiniband_hca_port_receive_data_bytes_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 37225401952885
		infiniband_hca_port_receive_data_bytes_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 9752484588300
		# HELP infiniband_hca_port_receive_errors_total Infiniband HCA port PortRcvErrors
		# TYPE infiniband_hca_port_receive_errors_total counter
		infiniband_hca_port_receive_errors_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_receive_errors_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_receive_packets_total Infiniband HCA port PortRcvPkts
		# TYPE infiniband_hca_port_receive_packets_total counter
		infiniband_hca_port_receive_packets_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 100583719365
		infiniband_hca_port_receive_packets_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 33038722564
		# HELP infiniband_hca_port_receive_remote_physical_errors_total Infiniband HCA port PortRcvRemotePhysicalErrors
		# TYPE infiniband_hca_port_receive_remote_physical_errors_total counter
		infiniband_hca_port_receive_remote_physical_errors_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_receive_remote_physical_errors_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_receive_switch_relay_errors_total Infiniband HCA port PortRcvSwitchRelayErrors
		# TYPE infiniband_hca_port_receive_switch_relay_errors_total counter
		infiniband_hca_port_receive_switch_relay_errors_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_receive_switch_relay_errors_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_symbol_error_total Infiniband HCA port SymbolErrorCounter
		# TYPE infiniband_hca_port_symbol_error_total counter
		infiniband_hca_port_symbol_error_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_symbol_error_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_transmit_constraint_errors_total Infiniband HCA port PortXmitConstraintErrors
		# TYPE infiniband_hca_port_transmit_constraint_errors_total counter
		infiniband_hca_port_transmit_constraint_errors_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_transmit_constraint_errors_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_transmit_data_bytes_total Infiniband HCA port PortXmitData
		# TYPE infiniband_hca_port_transmit_data_bytes_total counter
		infiniband_hca_port_transmit_data_bytes_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 37108676853855
		infiniband_hca_port_transmit_data_bytes_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 9049592493976
		# HELP infiniband_hca_port_transmit_discards_total Infiniband HCA port PortXmitDiscards
		# TYPE infiniband_hca_port_transmit_discards_total counter
		infiniband_hca_port_transmit_discards_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_transmit_discards_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_transmit_packets_total Infiniband HCA port PortXmitPkts
		# TYPE infiniband_hca_port_transmit_packets_total counter
		infiniband_hca_port_transmit_packets_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 96917117320
		infiniband_hca_port_transmit_packets_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 28825338611
		# HELP infiniband_hca_port_transmit_wait_total Infiniband HCA port PortXmitWait
		# TYPE infiniband_hca_port_transmit_wait_total counter
		infiniband_hca_port_transmit_wait_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_transmit_wait_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
		# HELP infiniband_hca_port_unicast_receive_packets_total Infiniband HCA port PortUnicastRcvPkts
		# TYPE infiniband_hca_port_unicast_receive_packets_total counter
		infiniband_hca_port_unicast_receive_packets_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 96851346228
		infiniband_hca_port_unicast_receive_packets_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 29306563974
		# HELP infiniband_hca_port_unicast_transmit_packets_total Infiniband HCA port PortUnicastXmitPkts
		# TYPE infiniband_hca_port_unicast_transmit_packets_total counter
		infiniband_hca_port_unicast_transmit_packets_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 96916572630
		infiniband_hca_port_unicast_transmit_packets_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 28824617123
		# HELP infiniband_hca_port_vl15_dropped_total Infiniband HCA port VL15Dropped
		# TYPE infiniband_hca_port_vl15_dropped_total counter
		infiniband_hca_port_vl15_dropped_total{guid="0x7cfe9003003b4b96",hca="o0002",port="1"} 0
		infiniband_hca_port_vl15_dropped_total{guid="0x7cfe9003003b4bde",hca="o0001",port="1"} 0
	`
	collector := NewHCACollector(&hcaDevices, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 47 {
		t.Errorf("Unexpected collection count %d, expected 47", val)
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
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestHCACollectorError(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	PerfqueryExec = func(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
		return "", fmt.Errorf("Error")
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="hca"} 2
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="hca"} 0
	`
	collector := NewHCACollector(&hcaDevices, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 3 {
		t.Errorf("Unexpected collection count %d, expected 3", val)
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
	PerfqueryExec = func(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
		return "", context.DeadlineExceeded
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="hca"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="hca"} 2
	`
	collector := NewHCACollector(&hcaDevices, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 3 {
		t.Errorf("Unexpected collection count %d, expected 3", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_hca_port_excessive_buffer_overrun_errors_total", "infiniband_hca_port_link_downed_total",
		"infiniband_hca_port_link_error_recovery_total", "infiniband_hca_port_local_link_integrity_errors_total",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
