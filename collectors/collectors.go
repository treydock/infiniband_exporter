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
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "infiniband"
)

var (
	useSudo         = kingpin.Flag("config.sudo", "Use sudo to execute IB commands").Default("true").Bool()
	execCommand     = exec.CommandContext
	collectDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil)
	collectErrors = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collect_errors"),
		"Number of errors that occurred during collection",
		[]string{"collector"}, nil)
	collecTimeouts = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collect_timeouts"),
		"Number of timeouts that occurred during collection",
		[]string{"collector"}, nil)
)

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
