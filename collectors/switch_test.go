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
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	switchDevices = []InfinibandDevice{
		{Type: "SW", LID: "2052", GUID: "0x506b4b03005c2740", Rate: (25 * 4 * 125000000), Name: "ib-i4l1s01",
			Uplinks: map[string]InfinibandUplink{
				"35": {Type: "CA", LID: "1432", PortNumber: "1", GUID: "0x506b4b0300cc02a6", Name: "p0001"},
			},
		},
		{Type: "SW", LID: "1719", GUID: "0x7cfe9003009ce5b0", Rate: (25 * 4 * 125000000), Name: "ib-i1l1s01",
			Uplinks: map[string]InfinibandUplink{
				"1":  {Type: "SW", LID: "1516", PortNumber: "1", GUID: "0x7cfe900300b07320", Name: "ib-i1l2s01"},
				"10": {Type: "CA", LID: "134", PortNumber: "1", GUID: "0x7cfe9003003b4bde", Name: "o0001"},
				"11": {Type: "CA", LID: "133", PortNumber: "1", GUID: "0x7cfe9003003b4b96", Name: "o0002"},
			},
		},
	}
)

func TestParseIBSWInfo(t *testing.T) {
	out, err := ReadFixture("ibswinfo", "test1")
	if err != nil {
		t.Fatal("Unable to read fixture")
	}
	data, err := parse_ibswinfo(out, log.NewNopLogger())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(data.PowerSupplies) != 2 {
		t.Errorf("Unexpected number of power supplies, got %d", len(data.PowerSupplies))
	}
	var psu0 SwitchPowerSupply
	for _, psu := range data.PowerSupplies {
		if psu.ID == "0" {
			psu0 = psu
			break
		}
	}
	if psu0.Status != "OK" {
		t.Errorf("Unexpected power supply status, got %s", psu0.Status)
	}
	if psu0.DCPower != "OK" {
		t.Errorf("Unexpected power supply dc power status, got %s", psu0.DCPower)
	}
	if psu0.FanStatus != "OK" {
		t.Errorf("Unexpected power supply fan status, got %s", psu0.FanStatus)
	}
	if psu0.PowerW != 72 {
		t.Errorf("Unexpected power supply watts, got %f", psu0.PowerW)
	}
	if data.Temp != 45 {
		t.Errorf("Unexpected temp, got %f", data.Temp)
	}
	if data.FanStatus != "OK" {
		t.Errorf("Unexpected fan status, got %s", data.FanStatus)
	}
	if len(data.Fans) != 8 {
		t.Errorf("Unexpected number of fans, got %d", len(data.Fans))
	}
	var fan1 SwitchFan
	for _, fan := range data.Fans {
		if fan.ID == "1" {
			fan1 = fan
			break
		}
	}
	if fan1.RPM != 8493 {
		t.Errorf("Unexpected fan RPM, got %f", fan1.RPM)
	}
}

func TestParseIBSWInfoErrors(t *testing.T) {
	tests := []string{
		"test-err1",
		"test-err2",
		"test-err3",
	}
	for i, test := range tests {
		out, err := ReadFixture("ibswinfo", test)
		if err != nil {
			t.Fatalf("Unable to read fixture %s", test)
		}
		_, err = parse_ibswinfo(out, log.NewNopLogger())
		if err == nil {
			t.Errorf("Expected an error for test %s(%d)", test, i)
		}
	}
}

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
		# HELP infiniband_switch_rate_bytes_per_second Infiniband switch rate
		# TYPE infiniband_switch_rate_bytes_per_second gauge
		infiniband_switch_rate_bytes_per_second{guid="0x506b4b03005c2740"} 1.25e+10
		infiniband_switch_rate_bytes_per_second{guid="0x7cfe9003009ce5b0"} 1.25e+10
		# HELP infiniband_switch_uplink_info Infiniband switch uplink information
		# TYPE infiniband_switch_uplink_info gauge
		infiniband_switch_uplink_info{guid="0x506b4b03005c2740",port="35",switch="ib-i4l1s01",uplink="p0001",uplink_guid="0x506b4b0300cc02a6",uplink_lid="1432",uplink_port="1",uplink_type="CA"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01",uplink="ib-i1l2s01",uplink_guid="0x7cfe900300b07320",uplink_lid="1516",uplink_port="1",uplink_type="SW"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="10",switch="ib-i1l1s01",uplink="o0001",uplink_guid="0x7cfe9003003b4bde",uplink_lid="134",uplink_port="1",uplink_type="CA"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="11",switch="ib-i1l1s01",uplink="o0002",uplink_guid="0x7cfe9003003b4b96",uplink_lid="133",uplink_port="1",uplink_type="CA"} 1
	`
	collector := NewSwitchCollector(&switchDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 77 {
		t.Errorf("Unexpected collection count %d, expected 77", val)
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
		"infiniband_switch_info", "infiniband_switch_rate_bytes_per_second", "infiniband_switch_uplink_info",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorFull(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{"--collector.switch.rcv-err-details", "--collector.switch.ibswinfo"}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, false, false)
	ibswinfoExec = func(lid string, ctx context.Context) (string, error) {
		if lid == "1719" {
			out, err := ReadFixture("ibswinfo", "test1")
			return out, err
		} else if lid == "2052" {
			out, err := ReadFixture("ibswinfo", "test2")
			return out, err
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
		# HELP infiniband_switch_fan_rpm Infiniband switch fan RPM
		# TYPE infiniband_switch_fan_rpm gauge
		infiniband_switch_fan_rpm{fan="1",guid="0x506b4b03005c2740"} 6125
		infiniband_switch_fan_rpm{fan="1",guid="0x7cfe9003009ce5b0"} 8493
		infiniband_switch_fan_rpm{fan="2",guid="0x506b4b03005c2740"} 5251
		infiniband_switch_fan_rpm{fan="2",guid="0x7cfe9003009ce5b0"} 7349
		infiniband_switch_fan_rpm{fan="3",guid="0x506b4b03005c2740"} 6013
		infiniband_switch_fan_rpm{fan="3",guid="0x7cfe9003009ce5b0"} 8441
		infiniband_switch_fan_rpm{fan="4",guid="0x506b4b03005c2740"} 5335
		infiniband_switch_fan_rpm{fan="4",guid="0x7cfe9003009ce5b0"} 7270
		infiniband_switch_fan_rpm{fan="5",guid="0x506b4b03005c2740"} 6068
		infiniband_switch_fan_rpm{fan="5",guid="0x7cfe9003009ce5b0"} 8337
		infiniband_switch_fan_rpm{fan="6",guid="0x506b4b03005c2740"} 5423
		infiniband_switch_fan_rpm{fan="6",guid="0x7cfe9003009ce5b0"} 7156
		infiniband_switch_fan_rpm{fan="7",guid="0x506b4b03005c2740"} 5854
		infiniband_switch_fan_rpm{fan="7",guid="0x7cfe9003009ce5b0"} 8441
		infiniband_switch_fan_rpm{fan="8",guid="0x506b4b03005c2740"} 5467
		infiniband_switch_fan_rpm{fan="8",guid="0x7cfe9003009ce5b0"} 7232
		infiniband_switch_fan_rpm{fan="9",guid="0x506b4b03005c2740"} 5906
		# HELP infiniband_switch_fan_status Infiniband switch fan status
		# TYPE infiniband_switch_fan_status gauge
		infiniband_switch_fan_status{guid="0x506b4b03005c2740",status="OK"} 1
		infiniband_switch_fan_status{guid="0x7cfe9003009ce5b0",status="OK"} 1
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
		# HELP infiniband_switch_power_supply_fan_status_info Infiniband switch power supply fan status
		# TYPE infiniband_switch_power_supply_fan_status_info gauge
		infiniband_switch_power_supply_fan_status_info{guid="0x506b4b03005c2740",psu="0",status="OK"} 1
		infiniband_switch_power_supply_fan_status_info{guid="0x506b4b03005c2740",psu="1",status="OK"} 1
		infiniband_switch_power_supply_fan_status_info{guid="0x7cfe9003009ce5b0",psu="0",status="OK"} 1
		infiniband_switch_power_supply_fan_status_info{guid="0x7cfe9003009ce5b0",psu="1",status="OK"} 1
		# HELP infiniband_switch_power_supply_dc_power_status_info Infiniband switch power supply DC power status
		# TYPE infiniband_switch_power_supply_dc_power_status_info gauge
		infiniband_switch_power_supply_dc_power_status_info{guid="0x506b4b03005c2740",psu="0",status="OK"} 1
		infiniband_switch_power_supply_dc_power_status_info{guid="0x506b4b03005c2740",psu="1",status="OK"} 1
		infiniband_switch_power_supply_dc_power_status_info{guid="0x7cfe9003009ce5b0",psu="0",status="OK"} 1
		infiniband_switch_power_supply_dc_power_status_info{guid="0x7cfe9003009ce5b0",psu="1",status="OK"} 1
		# HELP infiniband_switch_power_supply_status_info Infiniband switch power supply status
		# TYPE infiniband_switch_power_supply_status_info gauge
		infiniband_switch_power_supply_status_info{guid="0x506b4b03005c2740",psu="0",status="OK"} 1
		infiniband_switch_power_supply_status_info{guid="0x506b4b03005c2740",psu="1",status="OK"} 1
		infiniband_switch_power_supply_status_info{guid="0x7cfe9003009ce5b0",psu="0",status="OK"} 1
		infiniband_switch_power_supply_status_info{guid="0x7cfe9003009ce5b0",psu="1",status="OK"} 1
		# HELP infiniband_switch_power_supply_watts Infiniband switch power supply watts
		# TYPE infiniband_switch_power_supply_watts gauge
		infiniband_switch_power_supply_watts{guid="0x506b4b03005c2740",psu="0"} 154
		infiniband_switch_power_supply_watts{guid="0x506b4b03005c2740",psu="1"} 134
		infiniband_switch_power_supply_watts{guid="0x7cfe9003009ce5b0",psu="0"} 72
		infiniband_switch_power_supply_watts{guid="0x7cfe9003009ce5b0",psu="1"} 71
		# HELP infiniband_switch_rate_bytes_per_second Infiniband switch rate
		# TYPE infiniband_switch_rate_bytes_per_second gauge
		infiniband_switch_rate_bytes_per_second{guid="0x506b4b03005c2740"} 1.25e+10
		infiniband_switch_rate_bytes_per_second{guid="0x7cfe9003009ce5b0"} 1.25e+10
		# HELP infiniband_switch_temperature_celsius Infiniband switch temperature celsius
		# TYPE infiniband_switch_temperature_celsius gauge
		infiniband_switch_temperature_celsius{guid="0x506b4b03005c2740"} 53
		infiniband_switch_temperature_celsius{guid="0x7cfe9003009ce5b0"} 45
		# HELP infiniband_switch_uplink_info Infiniband switch uplink information
		# TYPE infiniband_switch_uplink_info gauge
		infiniband_switch_uplink_info{guid="0x506b4b03005c2740",port="35",switch="ib-i4l1s01",uplink="p0001",uplink_guid="0x506b4b0300cc02a6",uplink_lid="1432",uplink_port="1",uplink_type="CA"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01",uplink="ib-i1l2s01",uplink_guid="0x7cfe900300b07320",uplink_lid="1516",uplink_port="1",uplink_type="SW"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="10",switch="ib-i1l1s01",uplink="o0001",uplink_guid="0x7cfe9003003b4bde",uplink_lid="134",uplink_port="1",uplink_type="CA"} 1
		infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="11",switch="ib-i1l1s01",uplink="o0002",uplink_guid="0x7cfe9003003b4b96",uplink_lid="133",uplink_port="1",uplink_type="CA"} 1
	`
	collector := NewSwitchCollector(&switchDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 132 {
		t.Errorf("Unexpected collection count %d, expected 132", val)
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
		"infiniband_switch_info", "infiniband_switch_rate_bytes_per_second", "infiniband_switch_uplink_info",
		"infiniband_switch_power_supply_status_info", "infiniband_switch_power_supply_dc_power_status_info",
		"infiniband_switch_power_supply_fan_status_info", "infiniband_switch_power_supply_watts",
		"infiniband_switch_temperature_celsius", "infiniband_switch_fan_status", "infiniband_switch_fan_rpm",
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
	} else if val != 21 {
		t.Errorf("Unexpected collection count %d, expected 21", val)
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
		"infiniband_switch_info", "infiniband_switch_rate_bytes_per_second", "infiniband_switch_uplink_info",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorError(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{"--collector.switch.ibswinfo"}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, true, false)
	ibswinfoExec = func(lid string, ctx context.Context) (string, error) {
		var out string
		var err error
		if lid == "1719" {
			out, _ = ReadFixture("ibswinfo", "test-err1")
			err = nil
		} else if lid == "2052" {
			out = ""
			err = fmt.Errorf("Error")
		}
		return out, err
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 4
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 0
	`
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	collector := NewSwitchCollector(&switchDevices, false, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 11 {
		t.Errorf("Unexpected collection count %d, expected 11", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_switch_power_supply_status_info",
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
	} else if val != 12 {
		t.Errorf("Unexpected collection count %d, expected 12", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSwitchCollectorTimeout(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{"--collector.switch.ibswinfo"}); err != nil {
		t.Fatal(err)
	}
	SetPerfqueryExecs(t, false, true)
	ibswinfoExec = func(lid string, ctx context.Context) (string, error) {
		return "", context.DeadlineExceeded
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="switch"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="switch"} 4
	`
	collector := NewSwitchCollector(&switchDevices, false, log.NewNopLogger())
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 11 {
		t.Errorf("Unexpected collection count %d, expected 11", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_switch_port_excessive_buffer_overrun_errors_total", "infiniband_switch_port_link_downed_total",
		"infiniband_switch_port_link_error_recovery_total", "infiniband_switch_port_local_link_integrity_errors_total",
		"infiniband_switch_power_supply_status_info",
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestIBSWInfo(t *testing.T) {
	execCommand = fakeExecCommand
	mockedExitStatus = 0
	mockedStdout = "foo"
	defer func() { execCommand = exec.CommandContext }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := ibswinfo("1", ctx)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if out != mockedStdout {
		t.Errorf("Unexpected out: %s", out)
	}
}

func TestIBSWInfoError(t *testing.T) {
	execCommand = fakeExecCommand
	mockedExitStatus = 1
	mockedStdout = "foo"
	defer func() { execCommand = exec.CommandContext }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := ibswinfo("1", ctx)
	if err == nil {
		t.Errorf("Expected error")
	}
	if out != "" {
		t.Errorf("Unexpected out: %s", out)
	}
}

func TestIBSWInfoTimeout(t *testing.T) {
	execCommand = fakeExecCommand
	mockedExitStatus = 1
	mockedStdout = "foo"
	defer func() { execCommand = exec.CommandContext }()
	ctx, cancel := context.WithTimeout(context.Background(), 0*time.Second)
	defer cancel()
	out, err := ibswinfo("1", ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded")
	}
	if out != "" {
		t.Errorf("Unexpected out: %s", out)
	}
}
