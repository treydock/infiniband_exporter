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
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	mockedExitStatus = 0
	mockedStdout     string
	_, cancel        = context.WithTimeout(context.Background(), 5*time.Second)
	switchDevices    = []InfinibandDevice{
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

func SetIbnetdiscoverExec(t *testing.T, setErr bool, timeout bool) {
	IbnetdiscoverExec = func(ctx context.Context) (string, error) {
		if setErr {
			return "", fmt.Errorf("Error")
		}
		if timeout {
			return "", context.DeadlineExceeded
		}
		out, err := ReadFixture("ibnetdiscover", "test")
		if err != nil {
			t.Fatal(err.Error())
			return "", err
		}
		return out, nil
	}
}

func SetPerfqueryExecs(t *testing.T, setErr bool, timeout bool) {
	PerfqueryExec = func(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
		if setErr {
			return "", fmt.Errorf("Error")
		}
		if timeout {
			return "", context.DeadlineExceeded
		}
		var out string
		var err error
		if len(extraArgs) == 2 {
			out, err = ReadFixture("perfquery", guid)
			if err != nil {
				t.Fatal(err.Error())
				return "", err
			}
		} else {
			out, err = ReadFixture("perfquery-rcv-error", fmt.Sprintf("%s-%s", guid, port))
			if err != nil {
				t.Fatal(err.Error())
				return "", err
			}
		}
		return out, nil
	}
}

func fakeExecCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	es := strconv.Itoa(mockedExitStatus)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		"EXIT_STATUS=" + es}
	return cmd
}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	//nolint:staticcheck
	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

func setupGatherer(collector prometheus.Collector) prometheus.Gatherer {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)
	gatherers := prometheus.Gatherers{registry}
	return gatherers
}
