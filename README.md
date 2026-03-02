# Hetzner Storage Box Prometheus Exporter

This exporter collects statistics from all **Storage Boxes** in your **Hetzner Cloud project** and exposes them as Prometheus metrics.

## Docker Compose

```yaml
services:
  storage-box-exporter:
    image: ghcr.io/lukasmetzner/hetzner-storage-box-prometheus-exporter:v0.8.0 # x-releaser-pleaser-version
    ports:
      - 2112:2112
    environment:
      HETZNER_TOKEN: $HETZNER_TOKEN
      SCRAPE_INTERVAL: "30m"
```

* Metrics are available at: `http://<host>:2112/metrics`
* Requires a valid [Hetzner Cloud API token](https://docs.hetzner.cloud/reference/cloud), set via the `HETZNER_TOKEN` environment variable
* Configure the scrape interval at the Hetzner API with the environment variable `SCRAPE_INTERVAL`
    * Accepts values, which can be parsed by Go's [`time.ParseDuration()`](https://pkg.go.dev/time#ParseDuration)
    * Examples: `10s`, `5m`, `2h45m`
    * Default: `30m`

## Helm

The chart is published as an OCI artifact to GitHub Container Registry.

### Prerequisites

Create a Kubernetes Secret containing your Hetzner Cloud API token:

```bash
kubectl create secret generic hetzner --from-literal=token=<YOUR_HETZNER_TOKEN>
```

### Install

```bash
helm install storage-box-exporter oci://ghcr.io/lukasmetzner/charts/hetzner-storage-box-prometheus-exporter --version 0.8.0 # x-releaser-pleaser-version
```

### Values

| Key | Description | Default |
|-----|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `ghcr.io/lukasmetzner/hetzner-storage-box-prometheus-exporter` |
| `image.tag` | Image tag (defaults to chart `appVersion`) | `""` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `secret.name` | Name of existing Secret with the Hetzner API token | `"hetzner"` |
| `secret.key` | Key inside the Secret | `"token"` |
| `scrapeInterval` | Hetzner API scrape interval | `"30m"` |
| `resources` | Container resource requests/limits | `{}` |
| `serviceMonitor.enabled` | Deploy a ServiceMonitor for kube-prometheus-stack | `false` |
| `serviceMonitor.interval` | Prometheus scrape interval | `"1m"` |
| `serviceMonitor.scrapeTimeout` | Prometheus scrape timeout | `"30s"` |
| `serviceMonitor.additionalLabels` | Extra labels on the ServiceMonitor | `{}` |

> [!NOTE]
> `serviceMonitor.enabled: true` requires the [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) CRDs to be installed in your cluster.
