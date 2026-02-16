# Changelog

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
