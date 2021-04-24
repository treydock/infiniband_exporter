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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	ibnetdiscoverOut = `CA   134  1 0x7cfe9003003b4bde 4x EDR - SW  1719 10 0x7cfe9003009ce5b0 ( 'o0001 HCA-1' - 'ib-i1l1s01' )
CA   133  1 0x7cfe9003003b4b96 4x EDR - SW  1719 11 0x7cfe9003009ce5b0 ( 'o0002 HCA-1' - 'ib-i1l1s01' )
CA  1432  1 0x506b4b0300cc02a6 4x EDR - SW  2052 35 0x506b4b03005c2740 ( 'p0001 HCA-1' - 'ib-i4l1s01' )
SW  1719 10 0x7cfe9003009ce5b0 4x EDR - CA   134  1 0x7cfe9003003b4bde ( 'ib-i1l1s01' - 'o0001 HCA-1' )
SW  1719 11 0x7cfe9003009ce5b0 4x EDR - CA   133  1 0x7cfe9003003b4b96 ( 'ib-i1l1s01' - 'o0002 HCA-1' )
SW  1719  1 0x7cfe9003009ce5b0 4x EDR - SW  1516  1 0x7cfe900300b07320 ( 'ib-i1l1s01' - 'ib-i1l2s01' )
SW  2052 35 0x506b4b03005c2740 4x EDR - CA  1432  1 0x506b4b0300cc02a6 ( 'ib-i4l1s01' - 'p0001 HCA-1' )
SW  2052 37 0x506b4b03005c2740 4x ???                                    'ib-i4l1s01'
`
)

func TestIbnetdiscoverCollector(t *testing.T) {
	IbnetdiscoverExec = func(ctx context.Context) (string, error) {
		return ibnetdiscoverOut, nil
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="ibnetdiscover"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="ibnetdiscover"} 0
	`
	collector := NewIBNetDiscover(log.NewNopLogger())
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
	IbnetdiscoverExec = func(ctx context.Context) (string, error) {
		return "", fmt.Errorf("Error")
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="ibnetdiscover"} 1
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="ibnetdiscover"} 0
	`
	collector := NewIBNetDiscover(log.NewNopLogger())
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

func TestIbnetdiscoverCollectorTimeout(t *testing.T) {
	IbnetdiscoverExec = func(ctx context.Context) (string, error) {
		return "", context.DeadlineExceeded
	}
	expected := `
		# HELP infiniband_exporter_collect_errors Number of errors that occurred during collection
		# TYPE infiniband_exporter_collect_errors gauge
		infiniband_exporter_collect_errors{collector="ibnetdiscover"} 0
		# HELP infiniband_exporter_collect_timeouts Number of timeouts that occurred during collection
		# TYPE infiniband_exporter_collect_timeouts gauge
		infiniband_exporter_collect_timeouts{collector="ibnetdiscover"} 1
	`
	collector := NewIBNetDiscover(log.NewNopLogger())
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
		InfinibandDevice{Type: "CA", LID: "1432", GUID: "0x506b4b0300cc02a6", Rate: (25 * 4 * 125000000), Name: "p0001",
			Uplinks: map[string]InfinibandUplink{
				"1": InfinibandUplink{Type: "SW", LID: "2052", PortNumber: "35", GUID: "0x506b4b03005c2740", Name: "ib-i4l1s01"},
			},
		},
		InfinibandDevice{Type: "CA", LID: "133", GUID: "0x7cfe9003003b4b96", Rate: (25 * 4 * 125000000), Name: "o0002",
			Uplinks: map[string]InfinibandUplink{
				"1": InfinibandUplink{Type: "SW", LID: "1719", PortNumber: "11", GUID: "0x7cfe9003009ce5b0", Name: "ib-i1l1s01"},
			},
		},
		InfinibandDevice{Type: "CA", LID: "134", GUID: "0x7cfe9003003b4bde", Rate: (25 * 4 * 125000000), Name: "o0001",
			Uplinks: map[string]InfinibandUplink{
				"1": InfinibandUplink{Type: "SW", LID: "1719", PortNumber: "10", GUID: "0x7cfe9003009ce5b0", Name: "ib-i1l1s01"},
			},
		},
	}

	expectSwitches := []InfinibandDevice{
		InfinibandDevice{Type: "SW", LID: "2052", GUID: "0x506b4b03005c2740", Rate: (25 * 4 * 125000000), Name: "ib-i4l1s01",
			Uplinks: map[string]InfinibandUplink{
				"35": InfinibandUplink{Type: "CA", LID: "1432", PortNumber: "1", GUID: "0x506b4b0300cc02a6", Name: "p0001"},
			},
		},
		InfinibandDevice{Type: "SW", LID: "1719", GUID: "0x7cfe9003009ce5b0", Rate: (25 * 4 * 125000000), Name: "ib-i1l1s01",
			Uplinks: map[string]InfinibandUplink{
				"1":  InfinibandUplink{Type: "SW", LID: "1516", PortNumber: "1", GUID: "0x7cfe900300b07320", Name: "ib-i1l2s01"},
				"10": InfinibandUplink{Type: "CA", LID: "134", PortNumber: "1", GUID: "0x7cfe9003003b4bde", Name: "o0001"},
				"11": InfinibandUplink{Type: "CA", LID: "133", PortNumber: "1", GUID: "0x7cfe9003003b4b96", Name: "o0002"},
			},
		},
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	switches, hcas, err := ibnetdiscoverParse(ibnetdiscoverOut, logger)
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

func TestParseRate(t *testing.T) {
	tests := []struct {
		Width        string
		Rate         string
		ExpectedRate float64
	}{
		{Width: "4x", Rate: "SDR", ExpectedRate: 2 * 4 * 125000000},
		{Width: "4x", Rate: "DDR", ExpectedRate: 4 * 4 * 125000000},
		{Width: "4x", Rate: "QDR", ExpectedRate: 8 * 4 * 125000000},
		{Width: "4x", Rate: "FDR10", ExpectedRate: 10 * 4 * 125000000},
		{Width: "4x", Rate: "FDR", ExpectedRate: 14 * 4 * 125000000},
		{Width: "4x", Rate: "EDR", ExpectedRate: 25 * 4 * 125000000},
		{Width: "12x", Rate: "EDR", ExpectedRate: 25 * 12 * 125000000},
		{Width: "4x", Rate: "HDR", ExpectedRate: 50 * 4 * 125000000},
		{Width: "4x", Rate: "NDR", ExpectedRate: 100 * 4 * 125000000},
		{Width: "4x", Rate: "XDR", ExpectedRate: 250 * 4 * 125000000},
	}
	for i, test := range tests {
		rate, err := parseRate(test.Width, test.Rate)
		if err != nil {
			t.Errorf("Unexpected error in case %d: %s", i, err.Error())
			continue
		}
		if rate != test.ExpectedRate {
			t.Errorf("Unexpected rate in case %d:\nExpected: %v\nGot: %v", i, test.ExpectedRate, rate)
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
		_, err := parseRate(test.Width, test.Rate)
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
			ExpectedPortName: "o0001", ExpectedUplinkName: "ib-i1l1s01"},
		{Line: "SW  2052 35 0x506b4b03005c2740 4x EDR - CA  1432  1 0x506b4b0300cc02a6 ( 'ib-i4l1s01' - 'p0001 HCA-1' )",
			ExpectedPortName: "ib-i4l1s01", ExpectedUplinkName: "p0001"},
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
