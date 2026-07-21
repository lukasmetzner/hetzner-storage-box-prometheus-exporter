# Changelog

## [v0.9.1](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.9.1)

[Compare to previous version](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/compare/v0.9.0...v0.9.1)

### Bug Fixes

- **deps**: update module github.com/hetznercloud/hcloud-go/v2 to v2.37.0 (#25) ([30fc45c](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/30fc45ccb42f93f7ae9acf9c34f29d1f83773162))
- **deps**: update module github.com/hetznercloud/hcloud-go/v2 to v2.38.0 (#28) ([344d490](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/344d490a6be52b8e4536a4ac0dfb5dfea621d4ba))
- **deps**: update module github.com/hetznercloud/hcloud-go/v2 to v2.39.0 (#29) ([3ee785d](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/3ee785dc730a160c44bbc1664b7bb3e0aefec16e))
- **deps**: update module github.com/hetznercloud/hcloud-go/v2 to v2.40.0 (#30) ([859493a](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/859493aed22a1cfe7e9b5e753d8304fa2e151fad))
- **deps**: update module github.com/hetznercloud/hcloud-go/v2 to v2.41.1 (#31) ([af14eb9](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/af14eb9afac924fa990967553c9e0e462b44fc57))
- **deps**: update module github.com/hetznercloud/hcloud-go/v2 to v2.41.2 (#32) ([f2255a2](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/f2255a2dbe2dc1e4bfe21348e930a6c1c05f084f))
- **deps**: update module github.com/hetznercloud/hcloud-go/v2 to v2.43.0 (#33) ([d451bf7](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/d451bf7a7486ec7694bb0033452d2eb9dd1383a3))
- **deps**: update module github.com/hetznercloud/hcloud-go/v2 to v2.44.0 (#34) ([3fee792](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/3fee79258b71921f0e6d5723f8411918a018f96f))
- **deps**: update module github.com/hetznercloud/hcloud-go/v2 to v2.45.0 (#40) ([09ce694](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/09ce694011b4d6da218451018fb8214f72117fe8))
- **deps**: update module github.com/prometheus/client_golang to v1.24.0 (#41) ([cae5c8c](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/commit/cae5c8c93793e3c0883617679c9924e89559aa1f))

## [v0.9.0](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.9.0)

### Features

- drop subsystem name in Storage Box status

## [v0.8.0](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.8.0)

### Features

- add Helm chart

## [v0.7.0](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.7.0)

### Features

- rename HCLOUD_TOKEN to HETZNER_TOKEN

## [v0.6.0](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.6.0)

### Features

- rename stats_capacity to type_size

## [v0.5.0](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.5.0)

### Features

- add storage box capacity (#16)

## [v0.4.0](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.4.0)

### Features

- increase default interval to 30m
- stability improvements
- add /healthz and /readyz endpoints for container probes

### Bug Fixes

- log error without exiting
- reset gauge vectors before each scrape to prune stale metrics

## [v0.3.2](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.3.2)

### Bug Fixes

- exit on API errors

## [v0.3.1](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.3.1)

### Bug Fixes

- log error from metrics scraping

## [v0.3.0](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.3.0)

### Features

- storage box status metric

## [v0.2.0](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.2.0)

### Features

- goreleaser config

## [v0.1.0](https://github.com/lukasmetzner/hetzner-storage-box-prometheus-exporter/releases/tag/v0.1.0)

### Features

- init
