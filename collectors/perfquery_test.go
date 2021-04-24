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
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	perfqueryOut = `# Port extended counters: Lid 1719 port 1 (CapMask: 0x5300 CapMask2: 0x0000002)
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
	perfqueryOutRcvErrors = `# PortRcvErrorDetails counters: Lid 1432 port 1
PortSelect:......................1
CounterSelect:...................0x0000
PortLocalPhysicalErrors:.........0
PortMalformedPktErrors:..........0
PortBufferOverrunErrors:.........0
PortDLIDMappingErrors:...........0
PortVLMappingErrors:.............0
PortLoopingErrors:...............0
`
)

func TestPerfqueryParse(t *testing.T) {
	device := InfinibandDevice{GUID: "0x7cfe9003009ce5b0", Name: "test"}
	expected := []PerfQueryCounters{
		PerfQueryCounters{
			device:                       device,
			PortSelect:                   "1",
			PortXmitData:                 36298026860928,
			PortRcvData:                  12279028775751,
			PortXmitPkts:                 101733204203,
			PortRcvPkts:                  32262508468,
			PortUnicastXmitPkts:          101708165289,
			PortUnicastRcvPkts:           26677661727,
			PortMulticastXmitPkts:        25038914,
			PortMulticastRcvPkts:         5584846741,
			SymbolErrorCounter:           0,
			LinkErrorRecoveryCounter:     0,
			LinkDownedCounter:            0,
			PortRcvErrors:                0,
			PortRcvRemotePhysicalErrors:  0,
			PortRcvSwitchRelayErrors:     0,
			PortXmitDiscards:             0,
			PortXmitConstraintErrors:     0,
			PortRcvConstraintErrors:      0,
			LocalLinkIntegrityErrors:     0,
			ExcessiveBufferOverrunErrors: 0,
			VL15Dropped:                  0,
			PortXmitWait:                 22730501,
			QP1Dropped:                   0,
			PortLocalPhysicalErrors:      math.NaN(),
			PortMalformedPktErrors:       math.NaN(),
			PortBufferOverrunErrors:      math.NaN(),
			PortDLIDMappingErrors:        math.NaN(),
			PortVLMappingErrors:          math.NaN(),
			PortLoopingErrors:            math.NaN(),
		},
		PerfQueryCounters{
			device:                       device,
			PortSelect:                   "2",
			PortXmitData:                 26006570014026,
			PortRcvData:                  39078804993378,
			PortXmitPkts:                 122978948297,
			PortRcvPkts:                  93660802641,
			PortUnicastXmitPkts:          122978948297,
			PortUnicastRcvPkts:           93660802641,
			PortMulticastXmitPkts:        0,
			PortMulticastRcvPkts:         0,
			SymbolErrorCounter:           0,
			LinkErrorRecoveryCounter:     0,
			LinkDownedCounter:            0,
			PortRcvErrors:                0,
			PortRcvRemotePhysicalErrors:  0,
			PortRcvSwitchRelayErrors:     0,
			PortXmitDiscards:             0,
			PortXmitConstraintErrors:     0,
			PortRcvConstraintErrors:      0,
			LocalLinkIntegrityErrors:     0,
			ExcessiveBufferOverrunErrors: 0,
			VL15Dropped:                  0,
			PortXmitWait:                 36510964,
			QP1Dropped:                   0,
			PortLocalPhysicalErrors:      math.NaN(),
			PortMalformedPktErrors:       math.NaN(),
			PortBufferOverrunErrors:      math.NaN(),
			PortDLIDMappingErrors:        math.NaN(),
			PortVLMappingErrors:          math.NaN(),
			PortLoopingErrors:            math.NaN(),
		},
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	counters, err := perfqueryParse(device, perfqueryOut, logger)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if reflect.DeepEqual(expected, counters) {
		t.Errorf("Unexpected value\nExpected:\n%v\nGot:\n%v", expected, counters)
	}
}

func TestPerfqueryParseRcvErrorDetails(t *testing.T) {
	device := InfinibandDevice{GUID: "0x7cfe9003009ce5b0", Name: "test"}
	expected := []PerfQueryCounters{
		PerfQueryCounters{
			device:                       device,
			PortSelect:                   "1",
			PortXmitData:                 math.NaN(),
			PortRcvData:                  math.NaN(),
			PortXmitPkts:                 math.NaN(),
			PortRcvPkts:                  math.NaN(),
			PortUnicastXmitPkts:          math.NaN(),
			PortUnicastRcvPkts:           math.NaN(),
			PortMulticastXmitPkts:        math.NaN(),
			PortMulticastRcvPkts:         math.NaN(),
			SymbolErrorCounter:           math.NaN(),
			LinkErrorRecoveryCounter:     math.NaN(),
			LinkDownedCounter:            math.NaN(),
			PortRcvErrors:                math.NaN(),
			PortRcvRemotePhysicalErrors:  math.NaN(),
			PortRcvSwitchRelayErrors:     math.NaN(),
			PortXmitDiscards:             math.NaN(),
			PortXmitConstraintErrors:     math.NaN(),
			PortRcvConstraintErrors:      math.NaN(),
			LocalLinkIntegrityErrors:     math.NaN(),
			ExcessiveBufferOverrunErrors: math.NaN(),
			VL15Dropped:                  math.NaN(),
			PortXmitWait:                 math.NaN(),
			QP1Dropped:                   math.NaN(),
			PortLocalPhysicalErrors:      0,
			PortMalformedPktErrors:       0,
			PortBufferOverrunErrors:      0,
			PortDLIDMappingErrors:        0,
			PortVLMappingErrors:          0,
			PortLoopingErrors:            0,
		},
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	counters, err := perfqueryParse(device, perfqueryOutRcvErrors, logger)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if reflect.DeepEqual(expected, counters) {
		t.Errorf("Unexpected value\nExpected:\n%v\nGot:\n%v", expected, counters)
	}
}

func TestPerfqueryArgs(t *testing.T) {
	if _, err := kingpin.CommandLine.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	trueValue := true
	falseValue := false
	command, args := perfqueryArgs("0x00", "1", []string{"-l", "-x"})
	if command != "perfquery" {
		t.Errorf("Unexpected command, got: %s", command)
	}
	expectedArgs := []string{"-l", "-x", "-G", "0x00", "1"}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("Unexpected args\nExpected\n%v\nGot\n%v", expectedArgs, args)
	}
	useSudo = &trueValue
	command, args = perfqueryArgs("0x00", "1", []string{"-l", "-x"})
	if command != "sudo" {
		t.Errorf("Unexpected command, got: %s", command)
	}
	expectedArgs = []string{"perfquery", "-l", "-x", "-G", "0x00", "1"}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("Unexpected args\nExpected\n%v\nGot\n%v", expectedArgs, args)
	}
	useSudo = &falseValue
}

func TestPerfquery(t *testing.T) {
	execCommand = fakeExecCommand
	mockedExitStatus = 0
	mockedStdout = "foo"
	defer func() { execCommand = exec.CommandContext }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := perfquery("0x00", "1", []string{}, ctx)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if out != mockedStdout {
		t.Errorf("Unexpected out: %s", out)
	}
}

func TestPerfqueryError(t *testing.T) {
	execCommand = fakeExecCommand
	mockedExitStatus = 1
	mockedStdout = "foo"
	defer func() { execCommand = exec.CommandContext }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := perfquery("0x00", "1", []string{}, ctx)
	if err == nil {
		t.Errorf("Expected error")
	}
	if out != "" {
		t.Errorf("Unexpected out: %s", out)
	}
}

func TestPerfqueryTimeout(t *testing.T) {
	execCommand = fakeExecCommand
	mockedExitStatus = 1
	mockedStdout = "foo"
	defer func() { execCommand = exec.CommandContext }()
	ctx, cancel := context.WithTimeout(context.Background(), 0*time.Second)
	defer cancel()
	out, err := perfquery("0x00", "1", []string{}, ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded")
	}
	if out != "" {
		t.Errorf("Unexpected out: %s", out)
	}
}
