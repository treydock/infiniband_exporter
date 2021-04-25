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
	switchDevices = []InfinibandDevice{
		InfinibandDevice{GUID: "0x7cfe9003009ce5b0", Name: "ib-i1l1s01"},
		InfinibandDevice{GUID: "0x506b4b03005c2740", Name: "ib-i4l1s01"},
	}
	perfqueryOutSwitch1 = `# Port extended counters: Lid 1719 port 1 (CapMask: 0x5300 CapMask2: 0x0000002)
PortSelect:......................1
CounterSelect:...................0x0000
PortXmitData:....................36298026860928
PortRcvData:.....................12279028775751
PortXmitPkts:....................101733204203
PortRcvPkts:.....................32262508468
PortUnicastXmitPkts:.............101708165289
PortUnicastRcvPkts:..............26677661727
PortMulticastXmitPkts:...........25038914
PortMulticastRcvPkts:............5584846741
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
PortXmitWait:....................22730501
QP1Dropped:......................0
# Port extended counters: Lid 1719 port 2 (CapMask: 0x5300 CapMask2: 0x0000002)
PortSelect:......................2
CounterSelect:...................0x0000
PortXmitData:....................26006570014026
PortRcvData:.....................39078804993378
PortXmitPkts:....................122978948297
PortRcvPkts:.....................93660802641
PortUnicastXmitPkts:.............122978948297
PortUnicastRcvPkts:..............93660802641
PortMulticastXmitPkts:...........0
PortMulticastRcvPkts:............0
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
PortXmitWait:....................36510964
QP1Dropped:......................0
`
	perfqueryOutSwitch2 = `# Port extended counters: Lid 2052 port 1 (CapMask: 0x5300 CapMask2: 0x0000002)
PortSelect:......................1
CounterSelect:...................0x0000
PortXmitData:....................178791657177235
PortRcvData:.....................178762341961629
PortXmitPkts:....................393094651266
PortRcvPkts:.....................387654829341
PortUnicastXmitPkts:.............387471005571
PortUnicastRcvPkts:..............387648134400
PortMulticastXmitPkts:...........5623645694
PortMulticastRcvPkts:............6694940
CounterSelect2:..................0x00000000
SymbolErrorCounter:..............0
LinkErrorRecoveryCounter:........0
LinkDownedCounter:...............1
PortRcvErrors:...................0
PortRcvRemotePhysicalErrors:.....0
PortRcvSwitchRelayErrors:........7
PortXmitDiscards:................20046
PortXmitConstraintErrors:........0
PortRcvConstraintErrors:.........0
LocalLinkIntegrityErrors:........0
ExcessiveBufferOverrunErrors:....0
VL15Dropped:.....................0
PortXmitWait:....................41864608
QP1Dropped:......................0
`
	perfqueryRcvErrorSwitch1Port1 = `# PortRcvErrorDetails counters: Lid 1719 port 1
PortSelect:......................1
CounterSelect:...................0x0000
PortLocalPhysicalErrors:.........0
PortMalformedPktErrors:..........0
PortBufferOverrunErrors:.........0
PortDLIDMappingErrors:...........0
PortVLMappingErrors:.............0
PortLoopingErrors:...............0
`
	perfqueryRcvErrorSwitch1Port2 = `# PortRcvErrorDetails counters: Lid 1719 port 2
PortSelect:......................2
CounterSelect:...................0x0000
PortLocalPhysicalErrors:.........0
PortMalformedPktErrors:..........0
PortBufferOverrunErrors:.........0
PortDLIDMappingErrors:...........0
PortVLMappingErrors:.............0
PortLoopingErrors:...............0
`
	perfqueryRcvErrorSwitch2Port1 = `# PortRcvErrorDetails counters: Lid 2052 port 1
PortSelect:......................1
CounterSelect:...................0x0000
PortLocalPhysicalErrors:.........0
PortMalformedPktErrors:..........0
PortBufferOverrunErrors:.........0
PortDLIDMappingErrors:...........0
PortVLMappingErrors:.............0
PortLoopingErrors:...............0`
)

func TestSwitchCollector(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	PerfqueryExec = func(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
		if guid == "0x7cfe9003009ce5b0" {
			return perfqueryOutSwitch1, nil
		} else if guid == "0x506b4b03005c2740" {
			return perfqueryOutSwitch2, nil
		} else {
			return "", nil
		}
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 0
		# HELP infiniband_switch_port_excessive_buffer_overrun_errors_total Infiniband switch port ExcessiveBufferOverrunErrors
		# TYPE infiniband_switch_port_excessive_buffer_overrun_errors_total counter
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_link_downed_total Infiniband switch port LinkDownedCounter
		# TYPE infiniband_switch_port_link_downed_total counter
		infiniband_switch_port_link_downed_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 1
		infiniband_switch_port_link_downed_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_link_downed_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_link_error_recovery_total Infiniband switch port LinkErrorRecoveryCounter
		# TYPE infiniband_switch_port_link_error_recovery_total counter
		infiniband_switch_port_link_error_recovery_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_link_error_recovery_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_link_error_recovery_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_local_link_integrity_errors_total Infiniband switch port LocalLinkIntegrityErrors
		# TYPE infiniband_switch_port_local_link_integrity_errors_total counter
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_multicast_receive_packets_total Infiniband switch port PortMulticastRcvPkts
		# TYPE infiniband_switch_port_multicast_receive_packets_total counter
		infiniband_switch_port_multicast_receive_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 6694940
		infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 5584846741
		infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_multicast_transmit_packets_total Infiniband switch port PortMulticastXmitPkts
		# TYPE infiniband_switch_port_multicast_transmit_packets_total counter
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 5623645694
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 25038914
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_qp1_dropped_total Infiniband switch port QP1Dropped
		# TYPE infiniband_switch_port_qp1_dropped_total counter
		infiniband_switch_port_qp1_dropped_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_qp1_dropped_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_qp1_dropped_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_receive_constraint_errors_total Infiniband switch port PortRcvConstraintErrors
		# TYPE infiniband_switch_port_receive_constraint_errors_total counter
		infiniband_switch_port_receive_constraint_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_receive_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_receive_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_receive_data_bytes_total Infiniband switch port PortRcvData
		# TYPE infiniband_switch_port_receive_data_bytes_total counter
		infiniband_switch_port_receive_data_bytes_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 178762341961629
		infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 12279028775751
		infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 39078804993378
		# HELP infiniband_switch_port_receive_errors_total Infiniband switch port PortRcvErrors
		# TYPE infiniband_switch_port_receive_errors_total counter
		infiniband_switch_port_receive_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_receive_packets_total Infiniband switch port PortRcvPkts
		# TYPE infiniband_switch_port_receive_packets_total counter
		infiniband_switch_port_receive_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 387654829341
		infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 32262508468
		infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 93660802641
		# HELP infiniband_switch_port_receive_remote_physical_errors_total Infiniband switch port PortRcvRemotePhysicalErrors
		# TYPE infiniband_switch_port_receive_remote_physical_errors_total counter
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_receive_switch_relay_errors_total Infiniband switch port PortRcvSwitchRelayErrors
		# TYPE infiniband_switch_port_receive_switch_relay_errors_total counter
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 7
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_symbol_error_total Infiniband switch port SymbolErrorCounter
		# TYPE infiniband_switch_port_symbol_error_total counter
		infiniband_switch_port_symbol_error_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_symbol_error_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_symbol_error_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_transmit_constraint_errors_total Infiniband switch port PortXmitConstraintErrors
		# TYPE infiniband_switch_port_transmit_constraint_errors_total counter
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_transmit_data_bytes_total Infiniband switch port PortXmitData
		# TYPE infiniband_switch_port_transmit_data_bytes_total counter
		infiniband_switch_port_transmit_data_bytes_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 178791657177235
		infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 36298026860928
		infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 26006570014026
		# HELP infiniband_switch_port_transmit_discards_total Infiniband switch port PortXmitDiscards
		# TYPE infiniband_switch_port_transmit_discards_total counter
		infiniband_switch_port_transmit_discards_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 20046
		infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_transmit_packets_total Infiniband switch port PortXmitPkts
		# TYPE infiniband_switch_port_transmit_packets_total counter
		infiniband_switch_port_transmit_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 393094651266
		infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 101733204203
		infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 122978948297
		# HELP infiniband_switch_port_transmit_wait_total Infiniband switch port PortXmitWait
		# TYPE infiniband_switch_port_transmit_wait_total counter
		infiniband_switch_port_transmit_wait_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 41864608
		infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 22730501
		infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 36510964
		# HELP infiniband_switch_port_unicast_receive_packets_total Infiniband switch port PortUnicastRcvPkts
		# TYPE infiniband_switch_port_unicast_receive_packets_total counter
		infiniband_switch_port_unicast_receive_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 387648134400
		infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 26677661727
		infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 93660802641
		# HELP infiniband_switch_port_unicast_transmit_packets_total Infiniband switch port PortUnicastXmitPkts
		# TYPE infiniband_switch_port_unicast_transmit_packets_total counter
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 387471005571
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 101708165289
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 122978948297
		# HELP infiniband_switch_port_vl15_dropped_total Infiniband switch port VL15Dropped
		# TYPE infiniband_switch_port_vl15_dropped_total counter
		infiniband_switch_port_vl15_dropped_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
	`
	collector := NewSwitchCollector(&switchDevices, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 69 {
		t.Errorf("Unexpected collection count %d, expected 69", val)
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
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorFull(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{"--collector.switch.rcv-err-details"}); err != nil {
		t.Fatal(err)
	}
	PerfqueryExec = func(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
		if len(extraArgs) == 2 && guid == "0x7cfe9003009ce5b0" {
			return perfqueryOutSwitch1, nil
		} else if len(extraArgs) == 2 && guid == "0x506b4b03005c2740" {
			return perfqueryOutSwitch2, nil
		} else if guid == "0x7cfe9003009ce5b0" && port == "1" {
			return perfqueryRcvErrorSwitch1Port1, nil
		} else if guid == "0x7cfe9003009ce5b0" && port == "2" {
			return perfqueryRcvErrorSwitch1Port2, nil
		} else if guid == "0x506b4b03005c2740" && port == "1" {
			return perfqueryRcvErrorSwitch2Port1, nil
		} else {
			return "", nil
		}
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 0
		# HELP infiniband_switch_port_buffer_overrun_errors_total Infiniband switch port PortBufferOverrunErrors
		# TYPE infiniband_switch_port_buffer_overrun_errors_total counter
		infiniband_switch_port_buffer_overrun_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_dli_mapping_errors_total Infiniband switch port PortDLIDMappingErrors
		# TYPE infiniband_switch_port_dli_mapping_errors_total counter
		infiniband_switch_port_dli_mapping_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_dli_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_dli_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_excessive_buffer_overrun_errors_total Infiniband switch port ExcessiveBufferOverrunErrors
		# TYPE infiniband_switch_port_excessive_buffer_overrun_errors_total counter
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_excessive_buffer_overrun_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_link_downed_total Infiniband switch port LinkDownedCounter
		# TYPE infiniband_switch_port_link_downed_total counter
		infiniband_switch_port_link_downed_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 1
		infiniband_switch_port_link_downed_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_link_downed_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_link_error_recovery_total Infiniband switch port LinkErrorRecoveryCounter
		# TYPE infiniband_switch_port_link_error_recovery_total counter
		infiniband_switch_port_link_error_recovery_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_link_error_recovery_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_link_error_recovery_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_local_link_integrity_errors_total Infiniband switch port LocalLinkIntegrityErrors
		# TYPE infiniband_switch_port_local_link_integrity_errors_total counter
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_local_link_integrity_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_local_physical_errors_total Infiniband switch port PortLocalPhysicalErrors
		# TYPE infiniband_switch_port_local_physical_errors_total counter
		infiniband_switch_port_local_physical_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_local_physical_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_local_physical_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_looping_errors_total Infiniband switch port PortLoopingErrors
		# TYPE infiniband_switch_port_looping_errors_total counter
		infiniband_switch_port_looping_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_looping_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_looping_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_malformed_packet_errors_total Infiniband switch port PortMalformedPktErrors
		# TYPE infiniband_switch_port_malformed_packet_errors_total counter
		infiniband_switch_port_malformed_packet_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_malformed_packet_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_malformed_packet_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_multicast_receive_packets_total Infiniband switch port PortMulticastRcvPkts
		# TYPE infiniband_switch_port_multicast_receive_packets_total counter
		infiniband_switch_port_multicast_receive_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 6694940
		infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 5584846741
		infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_multicast_transmit_packets_total Infiniband switch port PortMulticastXmitPkts
		# TYPE infiniband_switch_port_multicast_transmit_packets_total counter
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 5623645694
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 25038914
		infiniband_switch_port_multicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_qp1_dropped_total Infiniband switch port QP1Dropped
		# TYPE infiniband_switch_port_qp1_dropped_total counter
		infiniband_switch_port_qp1_dropped_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_qp1_dropped_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_qp1_dropped_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_receive_constraint_errors_total Infiniband switch port PortRcvConstraintErrors
		# TYPE infiniband_switch_port_receive_constraint_errors_total counter
		infiniband_switch_port_receive_constraint_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_receive_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_receive_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_receive_data_bytes_total Infiniband switch port PortRcvData
		# TYPE infiniband_switch_port_receive_data_bytes_total counter
		infiniband_switch_port_receive_data_bytes_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 178762341961629
		infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 12279028775751
		infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 39078804993378
		# HELP infiniband_switch_port_receive_errors_total Infiniband switch port PortRcvErrors
		# TYPE infiniband_switch_port_receive_errors_total counter
		infiniband_switch_port_receive_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_receive_packets_total Infiniband switch port PortRcvPkts
		# TYPE infiniband_switch_port_receive_packets_total counter
		infiniband_switch_port_receive_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 387654829341
		infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 32262508468
		infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 93660802641
		# HELP infiniband_switch_port_receive_remote_physical_errors_total Infiniband switch port PortRcvRemotePhysicalErrors
		# TYPE infiniband_switch_port_receive_remote_physical_errors_total counter
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_receive_remote_physical_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_receive_switch_relay_errors_total Infiniband switch port PortRcvSwitchRelayErrors
		# TYPE infiniband_switch_port_receive_switch_relay_errors_total counter
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 7
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_receive_switch_relay_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_symbol_error_total Infiniband switch port SymbolErrorCounter
		# TYPE infiniband_switch_port_symbol_error_total counter
		infiniband_switch_port_symbol_error_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_symbol_error_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_symbol_error_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_transmit_constraint_errors_total Infiniband switch port PortXmitConstraintErrors
		# TYPE infiniband_switch_port_transmit_constraint_errors_total counter
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_transmit_constraint_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_transmit_data_bytes_total Infiniband switch port PortXmitData
		# TYPE infiniband_switch_port_transmit_data_bytes_total counter
		infiniband_switch_port_transmit_data_bytes_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 178791657177235
		infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 36298026860928
		infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 26006570014026
		# HELP infiniband_switch_port_transmit_discards_total Infiniband switch port PortXmitDiscards
		# TYPE infiniband_switch_port_transmit_discards_total counter
		infiniband_switch_port_transmit_discards_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 20046
		infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_transmit_packets_total Infiniband switch port PortXmitPkts
		# TYPE infiniband_switch_port_transmit_packets_total counter
		infiniband_switch_port_transmit_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 393094651266
		infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 101733204203
		infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 122978948297
		# HELP infiniband_switch_port_transmit_wait_total Infiniband switch port PortXmitWait
		# TYPE infiniband_switch_port_transmit_wait_total counter
		infiniband_switch_port_transmit_wait_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 41864608
		infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 22730501
		infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 36510964
		# HELP infiniband_switch_port_unicast_receive_packets_total Infiniband switch port PortUnicastRcvPkts
		# TYPE infiniband_switch_port_unicast_receive_packets_total counter
		infiniband_switch_port_unicast_receive_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 387648134400
		infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 26677661727
		infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 93660802641
		# HELP infiniband_switch_port_unicast_transmit_packets_total Infiniband switch port PortUnicastXmitPkts
		# TYPE infiniband_switch_port_unicast_transmit_packets_total counter
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 387471005571
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 101708165289
		infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 122978948297
		# HELP infiniband_switch_port_vl_mapping_errors_total Infiniband switch port PortVLMappingErrors
		# TYPE infiniband_switch_port_vl_mapping_errors_total counter
		infiniband_switch_port_vl_mapping_errors_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_vl_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_vl_mapping_errors_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
		# HELP infiniband_switch_port_vl15_dropped_total Infiniband switch port VL15Dropped
		# TYPE infiniband_switch_port_vl15_dropped_total counter
		infiniband_switch_port_vl15_dropped_total{guid="0x506b4b03005c2740",port="1",switch="ib-i4l1s01"} 0
		infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01"} 0
		infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="2",switch="ib-i1l1s01"} 0
	`
	collector := NewSwitchCollector(&switchDevices, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 87 {
		t.Errorf("Unexpected collection count %d, expected 87", val)
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
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorError(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	PerfqueryExec = func(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
		return "", fmt.Errorf("Error")
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 2
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 0
	`
	collector := NewSwitchCollector(&switchDevices, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 3 {
		t.Errorf("Unexpected collection count %d, expected 3", val)
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
	PerfqueryExec = func(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
		return "", context.DeadlineExceeded
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 2
	`
	collector := NewSwitchCollector(&switchDevices, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 3 {
		t.Errorf("Unexpected collection count %d, expected 3", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
