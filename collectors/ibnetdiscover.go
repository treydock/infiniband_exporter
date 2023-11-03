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

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	IbnetdiscoverExec    = ibnetdiscover
	ibnetdiscoverPath    = kingpin.Flag("ibnetdiscover.path", "Path to ibnetdiscover").Default("ibnetdiscover").String()
	nodeNameMap          = kingpin.Flag("ibnetdiscover.node-name-map", "Path to node name map file").Default("").String()
	ibnetdiscoverTimeout = kingpin.Flag("ibnetdiscover.timeout", "Timeout for ibnetdiscover execution").Default("20s").Duration()
	// IB Lane Rate Specification: {signaling rate, effective rate}, Gbps
	// 	https://en.wikipedia.org/wiki/InfiniBand#Performance
	laneRates = map[string][]float64{
		"SDR":   {2.5, 2},
		"DDR":   {5, 4},
		"QDR":   {10, 8},
		"FDR10": {10.3125, 10},
		"FDR":   {14.0625, 13.64},
		"EDR":   {25.78125, 25},
		"HDR":   {50, 50},
		"NDR":   {100, 100},
		"XDR":   {250, 250},
	}
)

type InfinibandDevice struct {
	Type    string
	LID     string
	GUID    string
	Rate    float64
	RawRate float64
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
		// check the last item, because name may have space so that it is split into multiple items
		name := items[len(items)-1]
		if strings.HasSuffix(name, `'`) && !isPairedQuotesName(name) {
			for i := len(items) - 2; i > 5; i-- {
				name = items[i] + name
				if isPairedQuotesName(name) {
					items = append(items[:i], name)
					break
				}
			}
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
		rawRate, effectiveRate, err := parseRate(items[4], items[5])
		if err != nil {
			level.Error(logger).Log("msg", "Unable to parse speed", "width", items[4], "rate", items[5], "type", device.Type, "guid", device.GUID)
			return nil, nil, err
		} else {
			device.Rate = effectiveRate
			device.RawRate = rawRate
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

func parseRate(width string, rateStr string) (float64, float64, error) {
	widthRe := regexp.MustCompile("[0-9]+")
	widthMatch := widthRe.FindAllString(width, 1)
	if len(widthMatch) != 1 {
		return 0, 0, fmt.Errorf("Unable to find match for %s: %v", width, widthMatch)
	}
	widthMultipler, _ := strconv.ParseFloat(widthMatch[0], 64)
	if laneRate, ok := laneRates[rateStr]; ok {
		baseRate := widthMultipler * math.Pow(1000, 3) / 8
		rawRate := laneRate[0] * baseRate
		effectiveRate := laneRate[1] * baseRate
		return rawRate, effectiveRate, nil
	}
	return 0, 0, fmt.Errorf("Unknown rate %s", rateStr)
}

func isPairedQuotesName(name string) bool {
	if name == `'` {
		return false
	}
	first, last := name[0], name[len(name)-1]
	return first == last && first == '\''
}

func parseNames(line string) (string, string, error) {
	re := regexp.MustCompile(`\( '(.+)' - '(.+)' \)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("Unable to extract names using regexp")
	}
	portName := strings.TrimSpace(matches[1])
	uplinkName := strings.TrimSpace(matches[2])
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
