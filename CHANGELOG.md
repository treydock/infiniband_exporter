## Unreleased

* Add `infiniband_exporter_last_execution` metric when exporter is run with `--exporter.runonce`

## 0.1.0 / 2021-07-03

* Add `--no-collector.hca.base-metrics` flag to disable collecting base HCA metrics
* Add `--no-collector.switch.base-metrics` flag to disable collecting base switch metrics
* When run with `--exporter.runonce`, the `collector` label will now have `-runonce` suffix to avoid conflicts with possible Prometheus scrape metrics

## 0.0.1 / 2021-04-27

### Changes

* Initial Release

