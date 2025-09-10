# Hetzner Storage Box Prometheus Exporter

Export Storage Box statistics to prometheus for all Storage Boxes in your Hetzner Cloud project. The metrics are exposed on port 2112 at `/metrics`. You need to set the environemnt variable `HCLOUD_TOKEN` to a valid [API token](https://docs.hetzner.cloud/reference/cloud).
