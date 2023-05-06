## 0.5.0 / 2023-05-06

* Update to Go 1.20 and update Go module dependencies (#13)

## 0.4.2 / 2022-12-07

* Rename infiniband_switch_fan_status to infiniband_switch_fan_status_info (#11)
* Include switch name with infiniband_switch_hardware_info (#11)

## 0.4.1 / 2022-12-07

* Ensure ibswinfo respects --sudo flag (#10)

## 0.4.0 / 2022-12-07

* Collect information from unmanaged switches using ibswinfo (BETA) (#9)

## 0.3.1 / 2022-08-24

* Handle switches with split mode enabled (#8)

## 0.3.0 / 2022-03-23

* Update to Go 1.17 and update Go module dependencies

## 0.2.0 / 2021-07-03

* Add `infiniband_exporter_last_execution` metric when exporter is run with `--exporter.runonce`

## 0.1.0 / 2021-07-03

* Add `--no-collector.hca.base-metrics` flag to disable collecting base HCA metrics
* Add `--no-collector.switch.base-metrics` flag to disable collecting base switch metrics
* When run with `--exporter.runonce`, the `collector` label will now have `-runonce` suffix to avoid conflicts with possible Prometheus scrape metrics

## 0.0.1 / 2021-04-27

### Changes

* Initial Release

