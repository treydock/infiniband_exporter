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

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/treydock/infiniband_exporter/collectors"
)

const (
	address = "localhost:19315"
)

var (
	outputPath     string
	expectedSwitch = `# HELP infiniband_switch_info Infiniband switch information
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
infiniband_switch_port_multicast_receive_packets_total{guid="0x506b4b03005c2740",port="1"} 6.69494e+06
infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 5.584846741e+09
infiniband_switch_port_multicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 0
# HELP infiniband_switch_port_multicast_transmit_packets_total Infiniband switch port PortMulticastXmitPkts
# TYPE infiniband_switch_port_multicast_transmit_packets_total counter
infiniband_switch_port_multicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1"} 5.623645694e+09
infiniband_switch_port_multicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 2.5038914e+07
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
infiniband_switch_port_receive_data_bytes_total{guid="0x506b4b03005c2740",port="1"} 1.78762341961629e+14
infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1"} 1.2279028775751e+13
infiniband_switch_port_receive_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2"} 3.9078804993378e+13
# HELP infiniband_switch_port_receive_errors_total Infiniband switch port PortRcvErrors
# TYPE infiniband_switch_port_receive_errors_total counter
infiniband_switch_port_receive_errors_total{guid="0x506b4b03005c2740",port="1"} 0
infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="1"} 0
infiniband_switch_port_receive_errors_total{guid="0x7cfe9003009ce5b0",port="2"} 0
# HELP infiniband_switch_port_receive_packets_total Infiniband switch port PortRcvPkts
# TYPE infiniband_switch_port_receive_packets_total counter
infiniband_switch_port_receive_packets_total{guid="0x506b4b03005c2740",port="1"} 3.87654829341e+11
infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 3.2262508468e+10
infiniband_switch_port_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 9.3660802641e+10
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
infiniband_switch_port_transmit_data_bytes_total{guid="0x506b4b03005c2740",port="1"} 1.78791657177235e+14
infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="1"} 3.6298026860928e+13
infiniband_switch_port_transmit_data_bytes_total{guid="0x7cfe9003009ce5b0",port="2"} 2.6006570014026e+13
# HELP infiniband_switch_port_transmit_discards_total Infiniband switch port PortXmitDiscards
# TYPE infiniband_switch_port_transmit_discards_total counter
infiniband_switch_port_transmit_discards_total{guid="0x506b4b03005c2740",port="1"} 20046
infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="1"} 0
infiniband_switch_port_transmit_discards_total{guid="0x7cfe9003009ce5b0",port="2"} 0
# HELP infiniband_switch_port_transmit_packets_total Infiniband switch port PortXmitPkts
# TYPE infiniband_switch_port_transmit_packets_total counter
infiniband_switch_port_transmit_packets_total{guid="0x506b4b03005c2740",port="1"} 3.93094651266e+11
infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 1.01733204203e+11
infiniband_switch_port_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 1.22978948297e+11
# HELP infiniband_switch_port_transmit_wait_total Infiniband switch port PortXmitWait
# TYPE infiniband_switch_port_transmit_wait_total counter
infiniband_switch_port_transmit_wait_total{guid="0x506b4b03005c2740",port="1"} 4.1864608e+07
infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="1"} 2.2730501e+07
infiniband_switch_port_transmit_wait_total{guid="0x7cfe9003009ce5b0",port="2"} 3.6510964e+07
# HELP infiniband_switch_port_unicast_receive_packets_total Infiniband switch port PortUnicastRcvPkts
# TYPE infiniband_switch_port_unicast_receive_packets_total counter
infiniband_switch_port_unicast_receive_packets_total{guid="0x506b4b03005c2740",port="1"} 3.876481344e+11
infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 2.6677661727e+10
infiniband_switch_port_unicast_receive_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 9.3660802641e+10
# HELP infiniband_switch_port_unicast_transmit_packets_total Infiniband switch port PortUnicastXmitPkts
# TYPE infiniband_switch_port_unicast_transmit_packets_total counter
infiniband_switch_port_unicast_transmit_packets_total{guid="0x506b4b03005c2740",port="1"} 3.87471005571e+11
infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="1"} 1.01708165289e+11
infiniband_switch_port_unicast_transmit_packets_total{guid="0x7cfe9003009ce5b0",port="2"} 1.22978948297e+11
# HELP infiniband_switch_port_vl15_dropped_total Infiniband switch port VL15Dropped
# TYPE infiniband_switch_port_vl15_dropped_total counter
infiniband_switch_port_vl15_dropped_total{guid="0x506b4b03005c2740",port="1"} 0
infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="1"} 0
infiniband_switch_port_vl15_dropped_total{guid="0x7cfe9003009ce5b0",port="2"} 0
# HELP infiniband_switch_rate_bytes_per_second Infiniband switch rate
# TYPE infiniband_switch_rate_bytes_per_second gauge
infiniband_switch_rate_bytes_per_second{guid="0x506b4b03005c2740"} 1.25e+10
infiniband_switch_rate_bytes_per_second{guid="0x7cfe9003009ce5b0"} 1.25e+10
# HELP infiniband_switch_raw_rate_bytes_per_second Infiniband switch raw rate
# TYPE infiniband_switch_raw_rate_bytes_per_second gauge
infiniband_switch_raw_rate_bytes_per_second{guid="0x506b4b03005c2740"} 1.2890625e+10
infiniband_switch_raw_rate_bytes_per_second{guid="0x7cfe9003009ce5b0"} 1.2890625e+10
# HELP infiniband_switch_uplink_info Infiniband switch uplink information
# TYPE infiniband_switch_uplink_info gauge
infiniband_switch_uplink_info{guid="0x506b4b03005c2740",port="35",switch="ib-i4l1s01",uplink="p0001 HCA-1",uplink_guid="0x506b4b0300cc02a6",uplink_lid="1432",uplink_port="1",uplink_type="CA"} 1
infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="1",switch="ib-i1l1s01",uplink="ib-i1l2s01",uplink_guid="0x7cfe900300b07320",uplink_lid="1516",uplink_port="1",uplink_type="SW"} 1
infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="10",switch="ib-i1l1s01",uplink="o0001 HCA-1",uplink_guid="0x7cfe9003003b4bde",uplink_lid="134",uplink_port="1",uplink_type="CA"} 1
infiniband_switch_uplink_info{guid="0x7cfe9003009ce5b0",port="11",switch="ib-i1l1s01",uplink="o0002 HCA-1",uplink_guid="0x7cfe9003003b4b96",uplink_lid="133",uplink_port="1",uplink_type="CA"} 1`
	expectedIbswinfo = `# HELP infiniband_switch_fan_rpm Infiniband switch fan RPM
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
# HELP infiniband_switch_fan_status_info Infiniband switch fan status
# TYPE infiniband_switch_fan_status_info gauge
infiniband_switch_fan_status_info{guid="0x506b4b03005c2740",status="OK"} 1
infiniband_switch_fan_status_info{guid="0x7cfe9003009ce5b0",status="ERROR"} 1
# HELP infiniband_switch_hardware_info Infiniband switch hardware info
# TYPE infiniband_switch_hardware_info gauge
infiniband_switch_hardware_info{firmware_version="11.2008.2102",guid="0x7cfe9003009ce5b0",part_number="MSB7790-ES2F",psid="MT_1880110032",serial_number="MT1943X00498",switch="ib-i1l1s01"} 1
infiniband_switch_hardware_info{firmware_version="27.2010.3118",guid="0x506b4b03005c2740",part_number="MQM8790-HS2F",psid="MT_0000000063",serial_number="MT2152T10239",switch="ib-i4l1s01"} 1
# HELP infiniband_switch_power_supply_dc_power_status_info Infiniband switch power supply DC power status
# TYPE infiniband_switch_power_supply_dc_power_status_info gauge
infiniband_switch_power_supply_dc_power_status_info{guid="0x506b4b03005c2740",psu="0",status="OK"} 1
infiniband_switch_power_supply_dc_power_status_info{guid="0x506b4b03005c2740",psu="1",status="OK"} 1
infiniband_switch_power_supply_dc_power_status_info{guid="0x7cfe9003009ce5b0",psu="0",status="OK"} 1
infiniband_switch_power_supply_dc_power_status_info{guid="0x7cfe9003009ce5b0",psu="1",status="OK"} 1
# HELP infiniband_switch_power_supply_fan_status_info Infiniband switch power supply fan status
# TYPE infiniband_switch_power_supply_fan_status_info gauge
infiniband_switch_power_supply_fan_status_info{guid="0x506b4b03005c2740",psu="0",status="OK"} 1
infiniband_switch_power_supply_fan_status_info{guid="0x506b4b03005c2740",psu="1",status="OK"} 1
infiniband_switch_power_supply_fan_status_info{guid="0x7cfe9003009ce5b0",psu="0",status="OK"} 1
infiniband_switch_power_supply_fan_status_info{guid="0x7cfe9003009ce5b0",psu="1",status="OK"} 1
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
# HELP infiniband_switch_temperature_celsius Infiniband switch temperature celsius
# TYPE infiniband_switch_temperature_celsius gauge
infiniband_switch_temperature_celsius{guid="0x506b4b03005c2740"} 53
infiniband_switch_temperature_celsius{guid="0x7cfe9003009ce5b0"} 45`
	expectedHCA = `# HELP infiniband_hca_info Infiniband HCA information
# TYPE infiniband_hca_info gauge
infiniband_hca_info{guid="0x506b4b0300cc02a6",hca="p0001 HCA-1",lid="1432"} 1
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
infiniband_hca_port_multicast_receive_packets_total{guid="0x7cfe9003003b4b96",port="1"} 3.732373137e+09
infiniband_hca_port_multicast_receive_packets_total{guid="0x7cfe9003003b4bde",port="1"} 3.732158589e+09
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
infiniband_hca_port_receive_data_bytes_total{guid="0x7cfe9003003b4b96",port="1"} 3.7225401952885e+13
infiniband_hca_port_receive_data_bytes_total{guid="0x7cfe9003003b4bde",port="1"} 9.7524845883e+12
# HELP infiniband_hca_port_receive_errors_total Infiniband HCA port PortRcvErrors
# TYPE infiniband_hca_port_receive_errors_total counter
infiniband_hca_port_receive_errors_total{guid="0x7cfe9003003b4b96",port="1"} 0
infiniband_hca_port_receive_errors_total{guid="0x7cfe9003003b4bde",port="1"} 0
# HELP infiniband_hca_port_receive_packets_total Infiniband HCA port PortRcvPkts
# TYPE infiniband_hca_port_receive_packets_total counter
infiniband_hca_port_receive_packets_total{guid="0x7cfe9003003b4b96",port="1"} 1.00583719365e+11
infiniband_hca_port_receive_packets_total{guid="0x7cfe9003003b4bde",port="1"} 3.3038722564e+10
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
infiniband_hca_port_transmit_data_bytes_total{guid="0x7cfe9003003b4b96",port="1"} 3.7108676853855e+13
infiniband_hca_port_transmit_data_bytes_total{guid="0x7cfe9003003b4bde",port="1"} 9.049592493976e+12
# HELP infiniband_hca_port_transmit_discards_total Infiniband HCA port PortXmitDiscards
# TYPE infiniband_hca_port_transmit_discards_total counter
infiniband_hca_port_transmit_discards_total{guid="0x7cfe9003003b4b96",port="1"} 0
infiniband_hca_port_transmit_discards_total{guid="0x7cfe9003003b4bde",port="1"} 0
# HELP infiniband_hca_port_transmit_packets_total Infiniband HCA port PortXmitPkts
# TYPE infiniband_hca_port_transmit_packets_total counter
infiniband_hca_port_transmit_packets_total{guid="0x7cfe9003003b4b96",port="1"} 9.691711732e+10
infiniband_hca_port_transmit_packets_total{guid="0x7cfe9003003b4bde",port="1"} 2.8825338611e+10
# HELP infiniband_hca_port_transmit_wait_total Infiniband HCA port PortXmitWait
# TYPE infiniband_hca_port_transmit_wait_total counter
infiniband_hca_port_transmit_wait_total{guid="0x7cfe9003003b4b96",port="1"} 0
infiniband_hca_port_transmit_wait_total{guid="0x7cfe9003003b4bde",port="1"} 0
# HELP infiniband_hca_port_unicast_receive_packets_total Infiniband HCA port PortUnicastRcvPkts
# TYPE infiniband_hca_port_unicast_receive_packets_total counter
infiniband_hca_port_unicast_receive_packets_total{guid="0x7cfe9003003b4b96",port="1"} 9.6851346228e+10
infiniband_hca_port_unicast_receive_packets_total{guid="0x7cfe9003003b4bde",port="1"} 2.9306563974e+10
# HELP infiniband_hca_port_unicast_transmit_packets_total Infiniband HCA port PortUnicastXmitPkts
# TYPE infiniband_hca_port_unicast_transmit_packets_total counter
infiniband_hca_port_unicast_transmit_packets_total{guid="0x7cfe9003003b4b96",port="1"} 9.691657263e+10
infiniband_hca_port_unicast_transmit_packets_total{guid="0x7cfe9003003b4bde",port="1"} 2.8824617123e+10
# HELP infiniband_hca_port_vl15_dropped_total Infiniband HCA port VL15Dropped
# TYPE infiniband_hca_port_vl15_dropped_total counter
infiniband_hca_port_vl15_dropped_total{guid="0x7cfe9003003b4b96",port="1"} 0
infiniband_hca_port_vl15_dropped_total{guid="0x7cfe9003003b4bde",port="1"} 0
# HELP infiniband_hca_rate_bytes_per_second Infiniband HCA rate
# TYPE infiniband_hca_rate_bytes_per_second gauge
infiniband_hca_rate_bytes_per_second{guid="0x506b4b0300cc02a6"} 1.25e+10
infiniband_hca_rate_bytes_per_second{guid="0x7cfe9003003b4b96"} 1.25e+10
infiniband_hca_rate_bytes_per_second{guid="0x7cfe9003003b4bde"} 1.25e+10
# HELP infiniband_hca_raw_rate_bytes_per_second Infiniband HCA raw rate
# TYPE infiniband_hca_raw_rate_bytes_per_second gauge
infiniband_hca_raw_rate_bytes_per_second{guid="0x506b4b0300cc02a6"} 1.2890625e+10
infiniband_hca_raw_rate_bytes_per_second{guid="0x7cfe9003003b4b96"} 1.2890625e+10
infiniband_hca_raw_rate_bytes_per_second{guid="0x7cfe9003003b4bde"} 1.2890625e+10
# HELP infiniband_hca_uplink_info Infiniband HCA uplink information
# TYPE infiniband_hca_uplink_info gauge
infiniband_hca_uplink_info{guid="0x506b4b0300cc02a6",hca="p0001 HCA-1",port="1",uplink="ib-i4l1s01",uplink_guid="0x506b4b03005c2740",uplink_lid="2052",uplink_port="35",uplink_type="SW"} 1
infiniband_hca_uplink_info{guid="0x7cfe9003003b4b96",hca="o0002 HCA-1",port="1",uplink="ib-i1l1s01",uplink_guid="0x7cfe9003009ce5b0",uplink_lid="1719",uplink_port="11",uplink_type="SW"} 1
infiniband_hca_uplink_info{guid="0x7cfe9003003b4bde",hca="o0001 HCA-1",port="1",uplink="ib-i1l1s01",uplink_guid="0x7cfe9003009ce5b0",uplink_lid="1719",uplink_port="10",uplink_type="SW"} 1`
	expectedSwitchNoError = `# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
# TYPE infiniband_exporter_collect_errors gauge
infiniband_exporter_collect_errors{collector="ibnetdiscover-runonce"} 0
infiniband_exporter_collect_errors{collector="switch-runonce"} 0
# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
# TYPE infiniband_exporter_collect_timeouts gauge
infiniband_exporter_collect_timeouts{collector="ibnetdiscover-runonce"} 0
infiniband_exporter_collect_timeouts{collector="switch-runonce"} 0`
	expectedFullNoError = `# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
# TYPE infiniband_exporter_collect_errors gauge
infiniband_exporter_collect_errors{collector="hca-runonce"} 0
infiniband_exporter_collect_errors{collector="ibnetdiscover-runonce"} 0
infiniband_exporter_collect_errors{collector="switch-runonce"} 0
# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
# TYPE infiniband_exporter_collect_timeouts gauge
infiniband_exporter_collect_timeouts{collector="hca-runonce"} 0
infiniband_exporter_collect_timeouts{collector="ibnetdiscover-runonce"} 0
infiniband_exporter_collect_timeouts{collector="switch-runonce"} 0`
	expectedIbnetdiscoverError = `# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
# TYPE infiniband_exporter_collect_errors gauge
infiniband_exporter_collect_errors{collector="ibnetdiscover-runonce"} 1
# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
# TYPE infiniband_exporter_collect_timeouts gauge
infiniband_exporter_collect_timeouts{collector="ibnetdiscover-runonce"} 0`
)

func TestMain(m *testing.M) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	collectors.IbnetdiscoverExec = func(ctx context.Context) (string, error) {
		out, err := collectors.ReadFixture("ibnetdiscover", "test")
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		return out, nil
	}
	collectors.PerfqueryExec = func(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
		out, err := collectors.ReadFixture("perfquery", guid)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		return out, nil
	}
	collectors.IbswinfoExec = func(lid string, ctx context.Context) (string, error) {
		if lid == "1719" {
			out, err := collectors.ReadFixture("ibswinfo", "test1")
			return out, err
		} else if lid == "2052" {
			out, err := collectors.ReadFixture("ibswinfo", "test2")
			return out, err
		} else {
			return "", nil
		}
	}
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestCollectToFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "output")
	if err != nil {
		os.Exit(1)
	}
	outputPath = tmpDir + "/output"
	defer os.RemoveAll(tmpDir)
	if _, err := kingpin.CommandLine.Parse([]string{fmt.Sprintf("--exporter.output=%s", outputPath), "--exporter.runonce"}); err != nil {
		t.Fatal(err)
	}
	err = run(log.NewNopLogger())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if !strings.Contains(string(content), expectedSwitch) {
		t.Errorf("Unexpected content:\nExpected:\n%s\nGot:\n%s", expectedSwitch, string(content))
	}
	if !strings.Contains(string(content), expectedSwitchNoError) {
		t.Errorf("Unexpected error content:\nExpected:\n%s\nGot:\n%s", expectedSwitchNoError, string(content))
	}
	if !strings.Contains(string(content), "infiniband_exporter_last_execution") {
		t.Errorf("Unexpected error content:\nExpected: infiniband_exporter_last_execution\nGot:\n%s", string(content))
	}
}

func TestCollect(t *testing.T) {
	var err error
	if _, err = kingpin.CommandLine.Parse([]string{fmt.Sprintf("--web.listen-address=%s", address)}); err != nil {
		t.Fatal(err)
	}
	go func() {
		err = run(log.NewNopLogger())
	}()
	if err != nil {
		t.Fatal(err)
	}
	body, err := queryExporter(metricsEndpoint)
	if err != nil {
		t.Fatalf("Unexpected error GET %s: %s", metricsEndpoint, err.Error())
	}
	if !strings.Contains(body, expectedSwitch) {
		t.Errorf("Unexpected body\nExpected:\n%s\nGot:\n%s\n", expectedSwitch, body)
	}
	// remove -runonce collector suffix
	runonceRe := regexp.MustCompile("-runonce")
	expectedSwitchNoError = runonceRe.ReplaceAllString(expectedSwitchNoError, "")
	if !strings.Contains(body, expectedSwitchNoError) {
		t.Errorf("Unexpected body\nExpected:\n%s\nGot:\n%s\n", expectedSwitchNoError, body)
	}
	if _, err = kingpin.CommandLine.Parse([]string{"--no-collector.switch", "--collector.ibswinfo", fmt.Sprintf("--web.listen-address=%s", address)}); err != nil {
		t.Fatal(err)
	}
	body, err = queryExporter(metricsEndpoint)
	if err != nil {
		t.Fatalf("Unexpected error GET %s: %s", metricsEndpoint, err.Error())
	}
	if !strings.Contains(body, expectedIbswinfo) {
		t.Errorf("Unexpected body\nExpected:\n%s\nGot:\n%s\n", expectedIbswinfo, body)
	}
	if _, err = kingpin.CommandLine.Parse([]string{"--collector.hca", fmt.Sprintf("--web.listen-address=%s", address)}); err != nil {
		t.Fatal(err)
	}
	body, err = queryExporter(metricsEndpoint)
	if err != nil {
		t.Fatalf("Unexpected error GET %s: %s", metricsEndpoint, err.Error())
	}
	if !strings.Contains(body, expectedHCA) {
		t.Errorf("Unexpected body\nExpected:\n%s\nGot:\n%s\n", expectedHCA, body)
	}
	expectedFullNoError = runonceRe.ReplaceAllString(expectedFullNoError, "")
	if !strings.Contains(body, expectedFullNoError) {
		t.Errorf("Unexpected body\nExpected:\n%s\nGot:\n%s\n", expectedFullNoError, body)
	}
	collectors.IbnetdiscoverExec = func(ctx context.Context) (string, error) {
		return "", fmt.Errorf("Error")
	}
	if _, err = kingpin.CommandLine.Parse([]string{"--web.disable-exporter-metrics", fmt.Sprintf("--web.listen-address=%s", address)}); err != nil {
		t.Fatal(err)
	}
	body, err = queryExporter(metricsEndpoint)
	if err != nil {
		t.Fatalf("Unexpected error GET %s: %s", metricsEndpoint, err.Error())
	}
	// Remove duration as can't mock value yet
	re := regexp.MustCompile(".*infiniband_exporter_collector_duration_seconds.*")
	body = re.ReplaceAllString(body, "")
	body = strings.TrimSpace(body)
	expectedIbnetdiscoverError = runonceRe.ReplaceAllString(expectedIbnetdiscoverError, "")
	if body != expectedIbnetdiscoverError {
		t.Errorf("Unexpected body\nExpected:\n%s\nGot:\n%s\n", expectedIbnetdiscoverError, body)
	}
}

func TestBaseURL(t *testing.T) {
	body, err := queryExporter("")
	if err != nil {
		t.Fatalf("Unexpected error GET base URL: %s", err.Error())
	}
	if !strings.Contains(body, metricsEndpoint) {
		t.Errorf("Unexpected body\nExpected: /metrics\nGot:\n%s\n", body)
	}
}

func queryExporter(path string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s%s", address, path))
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := resp.Body.Close(); err != nil {
		return "", err
	}
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		return "", fmt.Errorf("want /metrics status code %d, have %d. Body:\n%s", want, have, b)
	}
	return string(b), nil
}
