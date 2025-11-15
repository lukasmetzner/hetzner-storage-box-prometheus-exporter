package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var StorageBoxStatusList = []hcloud.StorageBoxStatus{
	hcloud.StorageBoxStatusInitializing,
	hcloud.StorageBoxStatusActive,
	hcloud.StorageBoxStatusLocked,
}

var (
	status = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "storage_box",
		Subsystem: "status",
		Name:      "status",
	}, []string{"storage-box", "status"})

	size = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "storage_box",
		Subsystem: "stats",
		Name:      "size",
	}, []string{"storage-box"})

	sizeData = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "storage_box",
		Subsystem: "stats",
		Name:      "size_data",
	}, []string{"storage-box"})

	sizeSnapshots = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "storage_box",
		Subsystem: "stats",
		Name:      "size_snapshots",
	}, []string{"storage-box"})
)

func scrapeMetrics(ctx context.Context, client *hcloud.Client) error {
	storageBoxes, err := client.StorageBox.All(ctx)
	if err != nil {
		return fmt.Errorf("error fetching storage boxes: %w", err)
	}

	for _, sbx := range storageBoxes {
		slog.Info("adding metrics", "storage-box-name", sbx.Name)

		for _, _status := range StorageBoxStatusList {
			if sbx.Status == _status {
				status.WithLabelValues(sbx.Name, string(_status)).Set(1.0)
			} else {
				status.WithLabelValues(sbx.Name, string(_status)).Set(0.0)
			}
		}

		size.With(prometheus.Labels{"storage-box": sbx.Name}).Set(float64(sbx.Stats.Size))
		sizeData.With(prometheus.Labels{"storage-box": sbx.Name}).Set(float64(sbx.Stats.SizeData))
		sizeSnapshots.With(prometheus.Labels{"storage-box": sbx.Name}).Set(float64(sbx.Stats.SizeSnapshots))
	}

	return nil
}

func run() error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	config, err := NewConfig()
	if err != nil {
		return err
	}

	opts := []hcloud.ClientOption{
		hcloud.WithToken(config.APIToken),
	}

	client := hcloud.NewClient(opts...)

	go func() {
		ctx := context.Background()
		for {
			if err := scrapeMetrics(ctx, client); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
			time.Sleep(config.ScrapeInterval)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(":2112", nil)
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
