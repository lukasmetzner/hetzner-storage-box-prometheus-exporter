package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	logger *slog.Logger

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

func recordMetrics(client *hcloud.Client) {
	go func() {
		ctx := context.Background()
		for {
			storageBoxes, err := client.StorageBox.All(ctx)
			if err != nil {
				fmt.Printf("error fetching storage boxes: %v\n", err)
				break
			}

			for _, sbx := range storageBoxes {
				logger.Info("adding metrics", "storage-box-name", sbx.Name)
				size.With(prometheus.Labels{"storage-box": sbx.Name}).Set(float64(sbx.Stats.Size))
				sizeData.With(prometheus.Labels{"storage-box": sbx.Name}).Set(float64(sbx.Stats.SizeData))
				sizeSnapshots.With(prometheus.Labels{"storage-box": sbx.Name}).Set(float64(sbx.Stats.SizeSnapshots))
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

func run() error {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	token, ok := os.LookupEnv("HCLOUD_TOKEN")
	if !ok {
		return fmt.Errorf("HCLOUD_TOKEN not set")
	}

	opts := []hcloud.ClientOption{
		hcloud.WithToken(token),
	}

	if debug, err := strconv.ParseBool(os.Getenv("HCLOUD_DEBUG")); err == nil && debug {
		opts = append(opts, hcloud.WithDebugWriter(os.Stderr))
	}

	client := hcloud.NewClient(opts...)

	recordMetrics(client)

	http.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(":2112", nil)
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
