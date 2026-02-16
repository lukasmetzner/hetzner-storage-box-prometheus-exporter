package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
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

	status.Reset()
	size.Reset()
	sizeData.Reset()
	sizeSnapshots.Reset()

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

const scrapeTimeout = 30 * time.Second

func run() error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := NewConfig()
	if err != nil {
		return err
	}

	opts := []hcloud.ClientOption{
		hcloud.WithToken(config.APIToken),
	}

	client := hcloud.NewClient(opts...)

	var ready atomic.Bool

	go func() {
		for {
			scrapeCtx, scrapeCancel := context.WithTimeout(ctx, scrapeTimeout)
			if err := scrapeMetrics(scrapeCtx, client); err != nil {
				slog.Error("scrape failed", "error", err)
			}
			scrapeCancel()
			ready.Store(true)

			select {
			case <-ctx.Done():
				slog.Info("stopping scrape loop")
				return
			case <-time.After(config.ScrapeInterval):
			}
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if ready.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, "ok")
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintln(w, "not ready")
	})

	server := &http.Server{
		Addr:         ":2112",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		slog.Info("shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}
	}()

	slog.Info("starting HTTP server", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
