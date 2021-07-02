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

package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gofrs/flock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/treydock/infiniband_exporter/collectors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	metricsEndpoint = "/metrics"
)

var (
	runOnce                = kingpin.Flag("exporter.runonce", "Run exporter once and write metrics to file").Default("false").Bool()
	output                 = kingpin.Flag("exporter.output", "Output file to write metrics to when using runonce").Default("").String()
	lockFile               = kingpin.Flag("exporter.lockfile", "Lock file path").Default("/tmp/infiniband_exporter.lock").String()
	listenAddress          = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9315").String()
	disableExporterMetrics = kingpin.Flag("web.disable-exporter-metrics", "Exclude metrics about the exporter (promhttp_*, process_*, go_*)").Default("false").Bool()
)

func setupGathers(runonce bool, logger log.Logger) prometheus.Gatherer {
	registry := prometheus.NewRegistry()

	ibnetdiscoverCollector := collectors.NewIBNetDiscover(runonce, logger)
	registry.MustRegister(ibnetdiscoverCollector)
	switches, hcas, err := ibnetdiscoverCollector.GetPorts()
	if err != nil {
		level.Error(logger).Log("msg", "Error collecting ports with ibnetdiscover", "err", err)
	} else {
		if *collectors.CollectSwitch {
			switchCollector := collectors.NewSwitchCollector(switches, runonce, logger)
			registry.MustRegister(switchCollector)
		}
		if *collectors.CollectHCA {
			hcaCollector := collectors.NewHCACollector(hcas, runonce, logger)
			registry.MustRegister(hcaCollector)
		}
	}

	gatherers := prometheus.Gatherers{registry}

	if !*disableExporterMetrics && !*runOnce {
		gatherers = append(gatherers, prometheus.DefaultGatherer)
	}
	return gatherers
}

func metricsHandler(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gatherers := setupGathers(false, logger)

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func writeMetrics(logger log.Logger) error {
	tmp, err := os.CreateTemp(filepath.Dir(*output), filepath.Base(*output))
	if err != nil {
		level.Error(logger).Log("msg", "Unable to create temporary file", "err", err)
		return err
	}
	defer os.Remove(tmp.Name())
	gatherers := setupGathers(true, logger)
	err = prometheus.WriteToTextfile(tmp.Name(), gatherers)
	if err != nil {
		level.Error(logger).Log("msg", "Error writing Prometheus metrics to file", "path", tmp.Name(), "err", err)
		return err
	}
	err = os.Rename(tmp.Name(), *output)
	if err != nil {
		level.Error(logger).Log("msg", "Error renaming temporary file to output", "tmp", tmp.Name(), "output", *output, "err", err)
		return err
	}
	return nil
}

func run(logger log.Logger) error {
	if *runOnce {
		if *output == "" {
			return fmt.Errorf("Must specify output path when using runonce mode")
		}
		fileLock := flock.New(*lockFile)
		unlocked, err := fileLock.TryLock()
		if err != nil {
			level.Error(logger).Log("msg", "Unable to obtain lock on lock file", "lockfile", *lockFile)
			return err
		}
		if !unlocked {
			return fmt.Errorf("Lock file %s is locked", *lockFile)
		}
		err = writeMetrics(logger)
		if err != nil {
			return err
		}
		return nil
	}
	level.Info(logger).Log("msg", "Starting infiniband_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	level.Info(logger).Log("msg", "Starting Server", "address", *listenAddress)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//nolint:errcheck
		w.Write([]byte(`<html>
             <head><title>InfiniBand Exporter</title></head>
             <body>
             <h1>InfiniBand Exporter</h1>
             <p><a href='` + metricsEndpoint + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	http.Handle(metricsEndpoint, metricsHandler(logger))
	err := http.ListenAndServe(*listenAddress, nil)
	return err
}

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("infiniband_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	err := run(logger)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
