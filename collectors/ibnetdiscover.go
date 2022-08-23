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
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	IbnetdiscoverExec    = ibnetdiscover
	ibnetdiscoverPath    = kingpin.Flag("ibnetdiscover.path", "Path to ibnetdiscover").Default("ibnetdiscover").String()
	nodeNameMap          = kingpin.Flag("ibnetdiscover.node-name-map", "Path to node name map file").Default("").String()
	ibnetdiscoverTimeout = kingpin.Flag("ibnetdiscover.timeout", "Timeout for ibnetdiscover execution").Default("20s").Duration()
	rates                = map[string]float64{
		"SDR":   2,
		"DDR":   4,
		"QDR":   8,
		"FDR10": 10,
		"FDR":   14,
		"EDR":   25,
		"HDR":   50,
		"NDR":   100,
		"XDR":   250,
	}
)

type InfinibandDevice struct {
	Type    string
	LID     string
	GUID    string
	Rate    float64
	Name    string
	Uplinks map[string]InfinibandUplink
}

type InfinibandUplink struct {
	Type       string
	LID        string
	PortNumber string
	GUID       string
	Name       string
}

type IBNetDiscover struct {
	timeoutMetric float64
	errorMetric   float64
	duration      float64
	logger        log.Logger
	collector     string
}

func NewIBNetDiscover(runonce bool, logger log.Logger) *IBNetDiscover {
	collector := "ibnetdiscover"
	if runonce {
		collector = "ibnetdiscover-runonce"
	}
	return &IBNetDiscover{
		logger:    log.With(logger, "collector", collector),
		collector: collector,
	}
}

func (ib *IBNetDiscover) GetPorts() (*[]InfinibandDevice, *[]InfinibandDevice, error) {
	collectTime := time.Now()
	switches, hcas, err := ib.collect()
	ib.duration = time.Since(collectTime).Seconds()
	if err == context.DeadlineExceeded {
		level.Error(ib.logger).Log("msg", "Timeout executing ibnetdiscover")
		ib.timeoutMetric = 1
	} else if err != nil {
		level.Error(ib.logger).Log("msg", "Error executing ibnetdiscover", "err", err)
		ib.errorMetric = 1
	}
	return switches, hcas, err
}

func (ib *IBNetDiscover) Describe(ch chan<- *prometheus.Desc) {
}

func (ib *IBNetDiscover) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(collectErrors, prometheus.GaugeValue, ib.errorMetric, ib.collector)
	ch <- prometheus.MustNewConstMetric(collecTimeouts, prometheus.GaugeValue, ib.timeoutMetric, ib.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, ib.duration, ib.collector)
	if strings.HasSuffix(ib.collector, "-runonce") {
		ch <- prometheus.MustNewConstMetric(lastExecution, prometheus.GaugeValue, float64(time.Now().Unix()), ib.collector)
	}
}

func (ib *IBNetDiscover) collect() (*[]InfinibandDevice, *[]InfinibandDevice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), *ibnetdiscoverTimeout)
	defer cancel()
	out, err := IbnetdiscoverExec(ctx)
	if err != nil {
		return nil, nil, err
	}
	switches, hcas, err := ibnetdiscoverParse(out, ib.logger)
	return switches, hcas, err
}

func ibnetdiscoverParse(out string, logger log.Logger) (*[]InfinibandDevice, *[]InfinibandDevice, error) {
	var switches, hcas []InfinibandDevice
	devices := make(map[string]InfinibandDevice)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		items := strings.Fields(line)
		if len(items) < 5 {
			level.Debug(logger).Log("msg", "Skipping line that is not connected", "line", line)
			continue
		}
		if items[5] == "???" {
			level.Debug(logger).Log("msg", "Skipping line that is not connected", "line", line)
			continue
		}
		if items[5] == "SDR" && len(items) == 7 {
			level.Debug(logger).Log("msg", "Skipping split mode port", "line", line)
			continue
		}
		guid := items[3]
		portNumber := items[2]
		var device InfinibandDevice
		var uplink InfinibandUplink
		if val, ok := devices[guid]; ok {
			device = val
		} else {
			device.Uplinks = make(map[string]InfinibandUplink)
		}
		device.Type = items[0]
		device.LID = items[1]
		device.GUID = guid
		rate, err := parseRate(items[4], items[5])
		if err != nil {
			level.Error(logger).Log("msg", "Unable to parse speed", "width", items[4], "rate", items[5], "type", device.Type, "guid", device.GUID)
			return nil, nil, err
		} else {
			device.Rate = rate
		}
		portName, uplinkName, err := parseNames(line)
		if err != nil {
			level.Error(logger).Log("msg", "Unable to parse names", "err", err, "type", device.Type, "guid", device.GUID, "line", line)
			return nil, nil, err
		}
		device.Name = portName
		uplink.Type = items[7]
		uplink.LID = items[8]
		uplink.PortNumber = items[9]
		uplink.GUID = items[10]
		uplink.Name = uplinkName
		device.Uplinks[portNumber] = uplink
		devices[guid] = device
	}
	deviceGUIDs := getDeviceGUIDs(devices)
	sort.Strings(deviceGUIDs)
	for _, guid := range deviceGUIDs {
		device := devices[guid]
		switch device.Type {
		case "CA":
			hcas = append(hcas, device)
		case "SW":
			switches = append(switches, device)
		}
	}
	return &switches, &hcas, nil
}

func parseRate(width string, rateStr string) (float64, error) {
	var rate float64
	widthRe := regexp.MustCompile("[0-9]+")
	widthMatch := widthRe.FindAllString(width, 1)
	if len(widthMatch) != 1 {
		return 0, fmt.Errorf("Unable to find match for %s: %v", width, widthMatch)
	}
	widthMultipler, _ := strconv.ParseFloat(widthMatch[0], 64)
	if baseRate, ok := rates[rateStr]; ok {
		rate = widthMultipler * baseRate * math.Pow(1000, 3) / 8
	} else {
		return 0, fmt.Errorf("Unknown rate %s", rateStr)
	}
	return rate, nil
}

func parseNames(line string) (string, string, error) {
	re := regexp.MustCompile(`\( '(.+)' - '(.+)' \)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("Unable to extract names using regexp")
	}
	portName := matches[1]
	uplinkName := matches[2]
	if strings.Contains(portName, " HCA") {
		portName = strings.Split(portName, " ")[0]
	}
	if strings.Contains(uplinkName, " HCA") {
		uplinkName = strings.Split(uplinkName, " ")[0]
	}
	return portName, uplinkName, nil
}

func getDeviceGUIDs(devices map[string]InfinibandDevice) []string {
	keys := make([]string, 0, len(devices))
	for key := range devices {
		keys = append(keys, key)
	}
	return keys
}

func getDevicePorts(uplinks map[string]InfinibandUplink) []string {
	keys := make([]string, 0, len(uplinks))
	for key := range uplinks {
		keys = append(keys, key)
	}
	return keys
}

func ibnetdiscoverArgs() (string, []string) {
	var command string
	var args []string
	if *useSudo {
		command = "sudo"
		args = []string{*ibnetdiscoverPath, "--ports"}
	} else {
		command = *ibnetdiscoverPath
		args = []string{"--ports"}
	}
	if *nodeNameMap != "" {
		args = append(args, "--node-name-map")
		args = append(args, *nodeNameMap)
	}
	return command, args
}

func ibnetdiscover(ctx context.Context) (string, error) {
	command, args := ibnetdiscoverArgs()
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
