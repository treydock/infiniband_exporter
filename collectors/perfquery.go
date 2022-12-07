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
	"bytes"
	"context"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	perfqueryPath    = kingpin.Flag("perfquery.path", "Path to perfquery").Default("perfquery").String()
	perfqueryTimeout = kingpin.Flag("perfquery.timeout", "Timeout for perfquery execution").Default("5s").Duration()
	maxConcurrent    = kingpin.Flag("perfquery.max-concurrent", "Max number of concurrent perfquery executions").Default("1").Int()
	PerfqueryExec    = perfquery
)

type PerfQueryCounters struct {
	device     InfinibandDevice
	PortSelect string
	// From -x / extended counters
	PortXmitData                 float64
	PortRcvData                  float64
	PortXmitPkts                 float64
	PortRcvPkts                  float64
	PortUnicastXmitPkts          float64
	PortUnicastRcvPkts           float64
	PortMulticastXmitPkts        float64
	PortMulticastRcvPkts         float64
	SymbolErrorCounter           float64
	LinkErrorRecoveryCounter     float64
	LinkDownedCounter            float64
	PortRcvErrors                float64
	PortRcvRemotePhysicalErrors  float64
	PortRcvSwitchRelayErrors     float64
	PortXmitDiscards             float64
	PortXmitConstraintErrors     float64
	PortRcvConstraintErrors      float64
	LocalLinkIntegrityErrors     float64
	ExcessiveBufferOverrunErrors float64
	VL15Dropped                  float64
	PortXmitWait                 float64
	QP1Dropped                   float64
	// From -E / PortRcvErrorDetails
	PortLocalPhysicalErrors float64
	PortMalformedPktErrors  float64
	PortBufferOverrunErrors float64
	PortDLIDMappingErrors   float64
	PortVLMappingErrors     float64
	PortLoopingErrors       float64
}

func initializeCounters(counters *PerfQueryCounters) {
	ps := reflect.ValueOf(counters)
	s := ps.Elem()
	v := reflect.TypeOf(*counters)
	for i := 0; i < v.NumField(); i++ {
		f := s.Field(i)
		if f.Kind() == reflect.Float64 {
			f.SetFloat(math.NaN())
		}
	}
}

func perfqueryParse(device InfinibandDevice, out string, logger log.Logger) ([]PerfQueryCounters, float64) {
	var counters []PerfQueryCounters
	var port string
	var errors float64
	portCounters := make(map[string]PerfQueryCounters)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		items := strings.Split(line, ":")
		if len(items) != 2 {
			level.Debug(logger).Log("msg", "Line has wrong number of elements, skipping", "line", line)
			continue
		}
		var counter PerfQueryCounters
		value := strings.Replace(items[1], ".", "", -1)
		if items[0] == "PortSelect" {
			port = value
		}
		if val, ok := portCounters[port]; ok {
			counter = val
		} else {
			initializeCounters(&counter)
			counter.device = device
		}
		ps := reflect.ValueOf(&counter)
		s := ps.Elem()
		f := s.FieldByName(items[0])
		if !f.IsValid() {
			level.Debug(logger).Log("msg", "Field not part of counters", "field", items[0])
			continue
		}
		if f.Kind() == reflect.String {
			f.SetString(value)
		} else if f.Kind() == reflect.Float64 {
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				level.Error(logger).Log("msg", "Unable to parse counter value", "err", err)
				errors++
				continue
			}
			f.SetFloat(val)
		}
		portCounters[port] = counter
	}
	for _, counter := range portCounters {
		counters = append(counters, counter)
	}
	return counters, errors
}

func perfqueryArgs(guid string, port string, extraArgs []string) (string, []string) {
	var command string
	var args []string
	if *useSudo {
		command = "sudo"
		args = []string{*perfqueryPath}
	} else {
		command = *perfqueryPath
	}
	args = append(args, extraArgs...)
	args = append(args, []string{"-G", guid}...)
	args = append(args, port)
	return command, args
}

func perfquery(guid string, port string, extraArgs []string, ctx context.Context) (string, error) {
	command, args := perfqueryArgs(guid, port, extraArgs)
	cmd := execCommand(ctx, command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return "", ctx.Err()
	} else if err != nil {
		return "", err
	}
	return out.String(), nil
}
