# Hetzner Storage Box Prometheus Exporter

This exporter collects statistics from all **Storage Boxes** in your **Hetzner Cloud project** and exposes them as Prometheus metrics.

* Metrics are available at: `http://<host>:2112/metrics`
* Requires a valid [Hetzner Cloud API token](https://docs.hetzner.cloud/reference/cloud), set via the `HCLOUD_TOKEN` environment variable
* Configure the scrape interval at the Hetzner API with the environment variable `SCRAPE_INTERVAL`
    * Accepts values, which can be parsed by Go's [`time.ParseDuration()`](https://pkg.go.dev/time#ParseDuration)
    * Examples: `10s`, `5m`, `2h45m`
