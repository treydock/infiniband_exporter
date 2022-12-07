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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "infiniband"
)

var (
	useSudo         = kingpin.Flag("sudo", "Use sudo to execute IB commands").Default("false").Bool()
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
	lastExecution = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "last_execution"),
		"Last execution time of exporter", []string{"collector"}, nil)
)

func ReadFixture(outputType string, name string) (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	if filepath.Base(dir) != "collectors" {
		dir = filepath.Join(dir, "collectors")
	}
	fixtureDir := filepath.Join(dir, "fixtures", outputType)
	fixture := filepath.Join(fixtureDir, fmt.Sprintf("%s.out", name))
	buffer, err := os.ReadFile(fixture)
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}
