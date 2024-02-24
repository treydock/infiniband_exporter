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
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

var (
	ibnetdiscoverBadRate = `CA   134  1 0x7cfe9003003b4bde 4x ZDR - SW  1719 10 0x7cfe9003009ce5b0 ( 'o0001 HCA-1' - 'ib-i1l1s01' )
SW  1719 10 0x7cfe9003009ce5b0 4x ZDR - CA   134  1 0x7cfe9003003b4bde ( 'ib-i1l1s01' - 'o0001 HCA-1' )`
	ibnetdiscoverBadName = `CA   134  1 0x7cfe9003003b4bde 4x EDR - SW  1719 10 0x7cfe9003009ce5b0 ( )
SW  1719 10 0x7cfe9003009ce5b0 4x EDR - CA   134  1 0x7cfe9003003b4bde ( )`
)

func TestIbnetdiscoverCollector(t *testing.T) {
	SetIbnetdiscoverExec(t, false, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="ibnetdiscover"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="ibnetdiscover"} 0
	`
	collector := NewIBNetDiscover(false, log.NewNopLogger())
	_, _, _ = collector.GetPorts()
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 3 {
		t.Errorf("Unexpected collection count %d, expected 3", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestIbnetdiscoverCollectorError(t *testing.T) {
	SetIbnetdiscoverExec(t, true, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="ibnetdiscover"} 1
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="ibnetdiscover"} 0
	`
	collector := NewIBNetDiscover(false, log.NewNopLogger())
	_, _, _ = collector.GetPorts()
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 3 {
		t.Errorf("Unexpected collection count %d, expected 3", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestIbnetdiscoverCollectorErrorRunonce(t *testing.T) {
	SetIbnetdiscoverExec(t, true, false)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="ibnetdiscover-runonce"} 1
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="ibnetdiscover-runonce"} 0
	`
	collector := NewIBNetDiscover(true, log.NewNopLogger())
	_, _, _ = collector.GetPorts()
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 4 {
		t.Errorf("Unexpected collection count %d, expected 4", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestIbnetdiscoverCollectorTimeout(t *testing.T) {
	SetIbnetdiscoverExec(t, false, true)
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="ibnetdiscover"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="ibnetdiscover"} 1
	`
	collector := NewIBNetDiscover(false, log.NewNopLogger())
	_, _, _ = collector.GetPorts()
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 3 {
		t.Errorf("Unexpected collection count %d, expected 3", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"infiniband_exporter_collect_errors", "infiniband_exporter_collect_timeouts"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestIbnetdiscoverParse(t *testing.T) {
	expectedHCAs := []InfinibandDevice{
		{Type: "CA", LID: "1432", GUID: "0x506b4b0300cc02a6", Rate: (25 * 4 * 125000000), RawRate: 1.2890625e+10, Name: "p0001 HCA-1",
			Uplinks: map[string]InfinibandUplink{
				"1": {Type: "SW", LID: "2052", PortNumber: "35", GUID: "0x506b4b03005c2740", Name: "ib-i4l1s01"},
			},
		},
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

	expectSwitches := []InfinibandDevice{
		{Type: "SW", LID: "2052", GUID: "0x506b4b03005c2740", Rate: (25 * 4 * 125000000), RawRate: 1.2890625e+10, Name: "ib-i4l1s01",
			Uplinks: map[string]InfinibandUplink{
				"35": {Type: "CA", LID: "1432", PortNumber: "1", GUID: "0x506b4b0300cc02a6", Name: "p0001 HCA-1"},
			},
		},
		{Type: "SW", LID: "1719", GUID: "0x7cfe9003009ce5b0", Rate: (25 * 4 * 125000000), RawRate: 1.2890625e+10, Name: "ib-i1l1s01",
			Uplinks: map[string]InfinibandUplink{
				"1":  {Type: "SW", LID: "1516", PortNumber: "1", GUID: "0x7cfe900300b07320", Name: "ib-i1l2s01"},
				"10": {Type: "CA", LID: "134", PortNumber: "1", GUID: "0x7cfe9003003b4bde", Name: "o0001 HCA-1"},
				"11": {Type: "CA", LID: "133", PortNumber: "1", GUID: "0x7cfe9003003b4b96", Name: "o0002 HCA-1"},
			},
		},
	}
	out, err := ReadFixture("ibnetdiscover", "test")
	if err != nil {
		t.Fatal("Unable to read fixture")
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	switches, hcas, err := ibnetdiscoverParse(out, logger)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if len(*hcas) != 3 {
		t.Errorf("Unexpected number of HCAs:\nExpected 3\nGot: %d", len(*hcas))
		return
	}
	if len(*switches) != 2 {
		t.Errorf("Unexpected number of switches:\nExpected 2\nGot: %d", len(*switches))
		return
	}
	for i, e := range expectedHCAs {
		if !reflect.DeepEqual((*hcas)[i], e) {
			t.Errorf("Unexpected value for HCA case %d:\nExpected: %v\nGot: %v", i, e, (*hcas)[i])
		}
	}
	for i, e := range expectSwitches {
		if !reflect.DeepEqual((*switches)[i], e) {
			t.Errorf("Unexpected value for switch case %d:\nExpected: %v\nGot: %v", i, e, (*switches)[i])
		}
	}
}

func TestIbnetdiscoverParse2(t *testing.T) {
	expectedHCAs := []InfinibandDevice{
		{Type: "CA", LID: "78", GUID: "0x946dae0300630bfe", Rate: 50 * 4 * 125000000, RawRate: 50 * 4 * 125000000, Name: "Mellanox Technologies Aggregation Node",
			Uplinks: map[string]InfinibandUplink{
				"1": {Type: "SW", LID: "51", PortNumber: "81", GUID: "0x946dae0300630bf6", Name: "5FB0405-leaf-IB01"},
			},
		},
		{Type: "CA", LID: "88", GUID: "0xb83fd20300da1138", Rate: 50 * 4 * 125000000, RawRate: 50 * 4 * 125000000, Name: "worker20 mlx5_3",
			Uplinks: map[string]InfinibandUplink{
				"1": {Type: "SW", LID: "51", PortNumber: "79", GUID: "0x946dae0300630bf6", Name: "5FB0405-leaf-IB01"},
			},
		},
	}

	expectSwitches := []InfinibandDevice{
		{Type: "SW", LID: "478", GUID: "0x0002c9020040f160", Rate: 8 * 4 * 125000000, RawRate: 10 * 4 * 125000000, Name: "Infiniscale-IV Mellanox Technologies",
			Uplinks: map[string]InfinibandUplink{},
		},
		{Type: "SW", LID: "9", GUID: "0x946dae030053ec1a", Rate: 50 * 4 * 125000000, RawRate: 50 * 4 * 125000000, Name: "5FB0406-spine-IB03",
			Uplinks: map[string]InfinibandUplink{
				"81": {Type: "CA", LID: "60", PortNumber: "1", GUID: "0x946dae0300630bfe", Name: "Mellanox Technologies Aggregation Node"},
			},
		},
	}
	out, err := ReadFixture("ibnetdiscover", "test2")
	if err != nil {
		t.Fatal("Unable to read fixture")
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	switches, hcas, err := ibnetdiscoverParse(out, logger)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if len(*hcas) != len(expectedHCAs) {
		t.Errorf("Unexpected number of HCAs:\nExpected %d\nGot: %d", len(expectedHCAs), len(*hcas))
		return
	}
	if len(*switches) != len(expectSwitches) {
		t.Errorf("Unexpected number of switches:\nExpected %d\nGot: %d", len(expectSwitches), len(*switches))
		return
	}
	for i, e := range expectedHCAs {
		if !reflect.DeepEqual((*hcas)[i], e) {
			t.Errorf("Unexpected value for HCA case %d:\nExpected: %v\nGot: %v", i, e, (*hcas)[i])
		}
	}
	for i, e := range expectSwitches {
		if !reflect.DeepEqual((*switches)[i], e) {
			t.Errorf("Unexpected value for switch case %d:\nExpected: %v\nGot: %v", i, e, (*switches)[i])
		}
	}
}

func TestIbnetdiscoverParse3(t *testing.T) {
	out, err := ReadFixture("ibnetdiscover", "test3")
	if err != nil {
		t.Fatal("Unable to read fixture")
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	_, _, err = ibnetdiscoverParse(out, logger)
	if err == nil || !strings.Contains(err.Error(), "Unable to extract names") {
		t.Errorf("Unexpected error: Unable to extract names")
		return
	}
}

func TestIbnetdiscoverParseErrors(t *testing.T) {
	tests := []struct {
		Input         string
		ExpectedError string
	}{
		{Input: ibnetdiscoverBadRate, ExpectedError: "Unknown rate ZDR"},
		{Input: ibnetdiscoverBadName, ExpectedError: "Unable to extract names using regexp"},
	}
	for i, test := range tests {
		_, _, err := ibnetdiscoverParse(test.Input, log.NewNopLogger())
		if err == nil {
			t.Errorf("Expected an error in case %d", i)
			continue
		}
		if err.Error() != test.ExpectedError {
			t.Errorf("Unexpected error in case %d:\nExpected: %v\nGot: %v", i, test.ExpectedError, err.Error())
		}
	}
}

func TestParseRate(t *testing.T) {
	tests := []struct {
		Width                 string
		Rate                  string
		ExpectedRawRate       float64
		ExpectedEffectiveRate float64
	}{
		{Width: "4x", Rate: "SDR", ExpectedRawRate: 2.5 * 4 * 125000000, ExpectedEffectiveRate: 2 * 4 * 125000000},
		{Width: "4x", Rate: "DDR", ExpectedRawRate: 5 * 4 * 125000000, ExpectedEffectiveRate: 4 * 4 * 125000000},
		{Width: "4x", Rate: "QDR", ExpectedRawRate: 10 * 4 * 125000000, ExpectedEffectiveRate: 8 * 4 * 125000000},
		{Width: "4x", Rate: "FDR10", ExpectedRawRate: 10.3125 * 4 * 125000000, ExpectedEffectiveRate: 10 * 4 * 125000000},
		{Width: "4x", Rate: "FDR", ExpectedRawRate: 14.0625 * 4 * 125000000, ExpectedEffectiveRate: 13.64 * 4 * 125000000},
		{Width: "4x", Rate: "EDR", ExpectedRawRate: 25.78125 * 4 * 125000000, ExpectedEffectiveRate: 25 * 4 * 125000000},
		{Width: "12x", Rate: "EDR", ExpectedRawRate: 25.78125 * 12 * 125000000, ExpectedEffectiveRate: 25 * 12 * 125000000},
		{Width: "4x", Rate: "HDR", ExpectedRawRate: 50 * 4 * 125000000, ExpectedEffectiveRate: 50 * 4 * 125000000},
		{Width: "4x", Rate: "NDR", ExpectedRawRate: 100 * 4 * 125000000, ExpectedEffectiveRate: 100 * 4 * 125000000},
		{Width: "4x", Rate: "XDR", ExpectedRawRate: 250 * 4 * 125000000, ExpectedEffectiveRate: 250 * 4 * 125000000},
	}
	for i, test := range tests {
		rawRate, effectiveRate, err := parseRate(test.Width, test.Rate)
		if err != nil {
			t.Errorf("Unexpected error in case %d: %s", i, err.Error())
			continue
		}
		if rawRate != test.ExpectedRawRate {
			t.Errorf("Unexpected raw rate in case %d:\nExpected: %v\nGot: %v", i, test.ExpectedRawRate, rawRate)
		}
		if effectiveRate != test.ExpectedEffectiveRate {
			t.Errorf("Unexpected effective rate in case %d:\nExpected: %v\nGot: %v", i, test.ExpectedEffectiveRate, effectiveRate)
		}
	}
}

func TestParseRateErrors(t *testing.T) {
	tests := []struct {
		Width         string
		Rate          string
		ExpectedError string
	}{
		{Width: "??", Rate: "EDR", ExpectedError: "Unable to find match for ??: []"},
		{Width: "4x", Rate: "ZDR", ExpectedError: "Unknown rate ZDR"},
	}
	for i, test := range tests {
		_, _, err := parseRate(test.Width, test.Rate)
		if err == nil {
			t.Errorf("Expected an error in case %d", i)
			continue
		}
		if err.Error() != test.ExpectedError {
			t.Errorf("Unexpected error in case %d:\nExpected: %v\nGot: %v", i, test.ExpectedError, err.Error())
		}
	}
}

func TestParseNames(t *testing.T) {
	tests := []struct {
		Line               string
		ExpectedPortName   string
		ExpectedUplinkName string
	}{
		{Line: "CA   134  1 0x7cfe9003003b4bde 4x EDR - SW  1719 10 0x7cfe9003009ce5b0 ( 'o0001 HCA-1' - 'ib-i1l1s01' )",
			ExpectedPortName: "o0001 HCA-1", ExpectedUplinkName: "ib-i1l1s01"},
		{Line: "SW  2052 35 0x506b4b03005c2740 4x EDR - CA  1432  1 0x506b4b0300cc02a6 ( 'ib-i4l1s01' - 'p0001 HCA-1' )",
			ExpectedPortName: "ib-i4l1s01", ExpectedUplinkName: "p0001 HCA-1"},
		{Line: "SW  1540 15 0x7cfe900300b07440 4x EDR - CA  1428  1 0x7cfe9003008dd6f8 ( 'SwitchIB Mellanox Technologies' - 'o0811 HCA-1' )",
			ExpectedPortName: "SwitchIB Mellanox Technologies", ExpectedUplinkName: "o0811 HCA-1"},
		{Line: "SW  1540  7 0x7cfe900300b07440 4x EDR - SW  1495 36 0x7cfe900300a1db20 ( 'SwitchIB Mellanox Technologies' - 'SwitchIB Mellanox Technologies' )",
			ExpectedPortName: "SwitchIB Mellanox Technologies", ExpectedUplinkName: "SwitchIB Mellanox Technologies"},
	}
	for i, test := range tests {
		portName, uplinkName, err := parseNames(test.Line)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
			continue
		}
		if portName != test.ExpectedPortName {
			t.Errorf("Unexpected port name in case %d:\nExpected: %v\nGot: %v", i, test.ExpectedPortName, portName)
		}
		if uplinkName != test.ExpectedUplinkName {
			t.Errorf("Unexpected uplink name in case %d:\nExpected: %v\nGot: %v", i, test.ExpectedUplinkName, uplinkName)
		}
	}
}

func TestParseNamesErrors(t *testing.T) {
	tests := []struct {
		Line          string
		ExpectedError string
	}{
		{Line: "SW  1540 10 0x7cfe900300b07440 4x EDR - CA    16  1 0x7cfe9003003b4b9a ( 'name' )",
			ExpectedError: "Unable to extract names using regexp"},
	}
	for i, test := range tests {
		_, _, err := parseNames(test.Line)
		if err == nil {
			t.Errorf("Expected an error in case %d", i)
			continue
		}
		if err.Error() != test.ExpectedError {
			t.Errorf("Unexpected error in case %d:\nExpected: %v\nGot: %v", i, test.ExpectedError, err.Error())
		}
	}
}

func TestGetDevicePorts(t *testing.T) {
	uplinks := map[string]InfinibandUplink{
		"10": {Type: "CA", LID: "134", PortNumber: "1", GUID: "0x7cfe9003003b4bde", Name: "o0001"},
		"11": {Type: "CA", LID: "133", PortNumber: "1", GUID: "0x7cfe9003003b4b96", Name: "o0002"},
	}
	expected := []string{"10", "11"}
	ports := getDevicePorts(uplinks)
	sort.Strings(expected)
	sort.Strings(ports)
	if !reflect.DeepEqual(ports, expected) {
		t.Errorf("Unexpected value for returned ports:\nExpected: %v\nGot: %v", expected, ports)
	}
}

func TestIbnetdiscoverArgs(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	trueValue := true
	falseValue := false
	nameMap := "/opt/foo"
	command, args := ibnetdiscoverArgs()
	if command != "ibnetdiscover" {
		t.Errorf("Unexpected command, got: %s", command)
	}
	if !reflect.DeepEqual(args, []string{"--ports"}) {
		t.Errorf("Unexpected args, got: %v", args)
	}
	useSudo = &trueValue
	command, args = ibnetdiscoverArgs()
	if command != "sudo" {
		t.Errorf("Unexpected command, got: %s", command)
	}
	if !reflect.DeepEqual(args, []string{"ibnetdiscover", "--ports"}) {
		t.Errorf("Unexpected args, got: %v", args)
	}
	nodeNameMap = &nameMap
	command, args = ibnetdiscoverArgs()
	if command != "sudo" {
		t.Errorf("Unexpected command, got: %s", command)
	}
	if !reflect.DeepEqual(args, []string{"ibnetdiscover", "--ports", "--node-name-map", "/opt/foo"}) {
		t.Errorf("Unexpected args, got: %v", args)
	}
	useSudo = &falseValue
}

func TestIbnetdiscover(t *testing.T) {
	execCommand = fakeExecCommand
	mockedExitStatus = 0
	mockedStdout = "foo"
	defer func() { execCommand = exec.CommandContext }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := ibnetdiscover(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if out != mockedStdout {
		t.Errorf("Unexpected out: %s", out)
	}
}

func TestIbnetdiscoverError(t *testing.T) {
	execCommand = fakeExecCommand
	mockedExitStatus = 1
	mockedStdout = "foo"
	defer func() { execCommand = exec.CommandContext }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := ibnetdiscover(ctx)
	if err == nil {
		t.Errorf("Expected error")
	}
	if out != "" {
		t.Errorf("Unexpected out: %s", out)
	}
}

func TestIibnetdiscoverTimeout(t *testing.T) {
	execCommand = fakeExecCommand
	mockedExitStatus = 1
	mockedStdout = "foo"
	defer func() { execCommand = exec.CommandContext }()
	ctx, cancel := context.WithTimeout(context.Background(), 0*time.Second)
	defer cancel()
	out, err := ibnetdiscover(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded")
	}
	if out != "" {
		t.Errorf("Unexpected out: %s", out)
	}
}
