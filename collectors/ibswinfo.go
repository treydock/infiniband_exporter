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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	CollectIbswinfo       = kingpin.Flag("collector.ibswinfo", "Enable ibswinfo data collection (BETA)").Default("false").Bool()
	ibswinfoPath          = kingpin.Flag("ibswinfo.path", "Path to ibswinfo").Default("ibswinfo").String()
	ibswinfoTimeout       = kingpin.Flag("ibswinfo.timeout", "Timeout for ibswinfo execution").Default("10s").Duration()
	ibswinfoMaxConcurrent = kingpin.Flag("ibswinfo.max-concurrent", "Max number of concurrent ibswinfo executions").Default("1").Int()
	IbswinfoExec          = ibswinfo
)

type IbswinfoCollector struct {
	devices              *[]InfinibandDevice
	logger               log.Logger
	collector            string
	HardwareInfo         *prometheus.Desc
	PowerSupplyStatus    *prometheus.Desc
	PowerSupplyDCPower   *prometheus.Desc
	PowerSupplyFanStatus *prometheus.Desc
	PowerSupplyWatts     *prometheus.Desc
	Temp                 *prometheus.Desc
	FanStatus            *prometheus.Desc
	FanRPM               *prometheus.Desc
}

type Ibswinfo struct {
	device          InfinibandDevice
	PartNumber      string
	SerialNumber    string
	PSID            string
	FirmwareVersion string
	PowerSupplies   []SwitchPowerSupply
	Temp            float64
	FanStatus       string
	Fans            []SwitchFan
}

type SwitchPowerSupply struct {
	ID        string
	Status    string
	DCPower   string
	FanStatus string
	PowerW    float64
}

type SwitchFan struct {
	ID  string
	RPM float64
}

func NewIbswinfoCollector(devices *[]InfinibandDevice, runonce bool, logger log.Logger) *IbswinfoCollector {
	collector := "ibswinfo"
	if runonce {
		collector = "ibswinfo-runonce"
	}
	return &IbswinfoCollector{
		devices:   devices,
		logger:    log.With(logger, "collector", collector),
		collector: collector,
		HardwareInfo: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "hardware_info"),
			"Infiniband switch hardware info", []string{"guid", "firmware_version", "psid", "part_number", "serial_number"}, nil),
		PowerSupplyStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "power_supply_status_info"),
			"Infiniband switch power supply status", []string{"guid", "psu", "status"}, nil),
		PowerSupplyDCPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "power_supply_dc_power_status_info"),
			"Infiniband switch power supply DC power status", []string{"guid", "psu", "status"}, nil),
		PowerSupplyFanStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "power_supply_fan_status_info"),
			"Infiniband switch power supply fan status", []string{"guid", "psu", "status"}, nil),
		PowerSupplyWatts: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "power_supply_watts"),
			"Infiniband switch power supply watts", []string{"guid", "psu"}, nil),
		Temp: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "temperature_celsius"),
			"Infiniband switch temperature celsius", []string{"guid"}, nil),
		FanStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "fan_status"),
			"Infiniband switch fan status", []string{"guid", "status"}, nil),
		FanRPM: prometheus.NewDesc(prometheus.BuildFQName(namespace, "switch", "fan_rpm"),
			"Infiniband switch fan RPM", []string{"guid", "fan"}, nil),
	}
}

func (s *IbswinfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.HardwareInfo
	ch <- s.PowerSupplyStatus
	ch <- s.PowerSupplyDCPower
	ch <- s.PowerSupplyFanStatus
	ch <- s.PowerSupplyWatts
	ch <- s.Temp
	ch <- s.FanStatus
	ch <- s.FanRPM
}

func (s *IbswinfoCollector) Collect(ch chan<- prometheus.Metric) {
	collectTime := time.Now()
	swinfos, errors, timeouts := s.collect()
	for _, swinfo := range swinfos {
		ch <- prometheus.MustNewConstMetric(s.HardwareInfo, prometheus.GaugeValue, 1, swinfo.device.GUID,
			swinfo.FirmwareVersion, swinfo.PSID, swinfo.PartNumber, swinfo.SerialNumber)
		for _, psu := range swinfo.PowerSupplies {
			ch <- prometheus.MustNewConstMetric(s.PowerSupplyStatus, prometheus.GaugeValue, 1, swinfo.device.GUID, psu.ID, psu.Status)
			ch <- prometheus.MustNewConstMetric(s.PowerSupplyDCPower, prometheus.GaugeValue, 1, swinfo.device.GUID, psu.ID, psu.DCPower)
			ch <- prometheus.MustNewConstMetric(s.PowerSupplyFanStatus, prometheus.GaugeValue, 1, swinfo.device.GUID, psu.ID, psu.FanStatus)
			ch <- prometheus.MustNewConstMetric(s.PowerSupplyWatts, prometheus.GaugeValue, psu.PowerW, swinfo.device.GUID, psu.ID)
		}
		ch <- prometheus.MustNewConstMetric(s.Temp, prometheus.GaugeValue, swinfo.Temp, swinfo.device.GUID)
		ch <- prometheus.MustNewConstMetric(s.FanStatus, prometheus.GaugeValue, 1, swinfo.device.GUID, swinfo.FanStatus)
		for _, fan := range swinfo.Fans {
			ch <- prometheus.MustNewConstMetric(s.FanRPM, prometheus.GaugeValue, fan.RPM, swinfo.device.GUID, fan.ID)
		}
	}
	ch <- prometheus.MustNewConstMetric(collectErrors, prometheus.GaugeValue, errors, s.collector)
	ch <- prometheus.MustNewConstMetric(collecTimeouts, prometheus.GaugeValue, timeouts, s.collector)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), s.collector)
	if strings.HasSuffix(s.collector, "-runonce") {
		ch <- prometheus.MustNewConstMetric(lastExecution, prometheus.GaugeValue, float64(time.Now().Unix()), s.collector)
	}
}

func (s *IbswinfoCollector) collect() ([]Ibswinfo, float64, float64) {
	var ibswinfos []Ibswinfo
	var ibswinfosLock sync.Mutex
	var errors, timeouts float64
	limit := make(chan int, *ibswinfoMaxConcurrent)
	wg := &sync.WaitGroup{}
	level.Debug(s.logger).Log("msg", "Collecting ibswinfo on devices", "count", len(*s.devices))
	for _, device := range *s.devices {
		limit <- 1
		wg.Add(1)
		go func(device InfinibandDevice) {
			defer wg.Done()
			ctxibswinfo, cancelibswinfo := context.WithTimeout(context.Background(), *ibswinfoTimeout)
			defer cancelibswinfo()
			level.Debug(s.logger).Log("msg", "Run ibswinfo", "lid", device.LID)
			ibswinfoOut, ibswinfoErr := IbswinfoExec(device.LID, ctxibswinfo)
			if ibswinfoErr == context.DeadlineExceeded {
				level.Error(s.logger).Log("msg", "Timeout collecting ibswinfo data", "guid", device.GUID, "lid", device.LID)
				timeouts++
			} else if ibswinfoErr != nil {
				level.Error(s.logger).Log("msg", "Error collecting ibswinfo data", "err", fmt.Sprintf("%s:%s", ibswinfoErr, ibswinfoOut), "guid", device.GUID, "lid", device.LID)
				errors++
			}
			if ibswinfoErr == nil {
				ibswinfoData, err := parse_ibswinfo(ibswinfoOut, s.logger)
				if err != nil {
					level.Error(s.logger).Log("msg", "Error parsing ibswinfo output", "guid", device.GUID, "lid", device.LID)
					errors++
				} else {
					ibswinfoData.device = device
					ibswinfosLock.Lock()
					ibswinfos = append(ibswinfos, ibswinfoData)
					ibswinfosLock.Unlock()
				}
			}
			<-limit
		}(device)
	}
	wg.Wait()
	close(limit)
	return ibswinfos, errors, timeouts
}

func ibswinfoArgs(lid string) (string, []string) {
	var command string
	var args []string
	if *useSudo {
		command = "sudo"
		args = []string{*ibswinfoPath}
	} else {
		command = *ibswinfoPath
	}
	args = append(args, []string{"-d", fmt.Sprintf("lid-%s", lid)}...)
	return command, args
}

func ibswinfo(lid string, ctx context.Context) (string, error) {
	command, args := ibswinfoArgs(lid)
	cmd := execCommand(ctx, command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return "", ctx.Err()
	} else if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}

func parse_ibswinfo(out string, logger log.Logger) (Ibswinfo, error) {
	var data Ibswinfo
	lines := strings.Split(out, "\n")
	psus := make(map[string]SwitchPowerSupply)
	var powerSupplies []SwitchPowerSupply
	var fans []SwitchFan
	var psuID string
	rePSU := regexp.MustCompile(`PSU([0-9]) status`)
	reFan := regexp.MustCompile(`fan#([0-9]+)`)
	for _, line := range lines {
		l := strings.Split(line, "|")
		if len(l) != 2 {
			continue
		}
		key := strings.TrimSpace(l[0])
		value := strings.TrimSpace(l[1])
		switch key {
		case "part number":
			data.PartNumber = value
		case "serial number":
			data.SerialNumber = value
		case "PSID":
			data.PSID = value
		case "firmware version":
			data.FirmwareVersion = value
		}
		matchesPSU := rePSU.FindStringSubmatch(key)
		var psu SwitchPowerSupply
		if psuID != "" {
			if p, ok := psus[psuID]; ok {
				psu = p
			}
		}
		if len(matchesPSU) == 2 {
			psuID = matchesPSU[1]
			psu.Status = value
		}
		if key == "DC power" {
			psu.DCPower = value
		}
		if psuID != "" && key == "fan status" {
			psu.FanStatus = value
		}
		if key == "power (W)" {
			powerW, err := strconv.ParseFloat(value, 64)
			if err == nil {
				psu.PowerW = powerW
			} else {
				level.Error(logger).Log("msg", "Unable to parse power (W)", "err", err, "value", value)
				return Ibswinfo{}, err
			}
		}
		if psuID != "" {
			psus[psuID] = psu
		}
		if key == "power (W)" {
			psuID = ""
		}
		if key == "temperature (C)" {
			temp, err := strconv.ParseFloat(value, 64)
			if err == nil {
				data.Temp = temp
			} else {
				level.Error(logger).Log("msg", "Unable to parse temperature (C)", "err", err, "value", value)
				return Ibswinfo{}, err
			}
		}
		if psuID == "" && key == "fan status" {
			data.FanStatus = value
		}
		matchesFan := reFan.FindStringSubmatch(key)
		if len(matchesFan) == 2 {
			rpm, err := strconv.ParseFloat(value, 64)
			if err == nil {
				fan := SwitchFan{
					ID:  matchesFan[1],
					RPM: rpm,
				}
				fans = append(fans, fan)
			} else {
				level.Error(logger).Log("msg", "Unable to parse fan RPM", "err", err, "value", value)
				return Ibswinfo{}, err
			}
		}
	}
	for id, psu := range psus {
		psu.ID = id
		powerSupplies = append(powerSupplies, psu)
	}
	data.PowerSupplies = powerSupplies
	data.Fans = fans
	return data, nil
}
