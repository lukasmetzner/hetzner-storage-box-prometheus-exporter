package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// newTestClient returns an hcloud client whose Storage Box API is served by
// handler. Retries are disabled so failure cases do not stall the test.
func newTestClient(t *testing.T, handler http.Handler) *hcloud.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return hcloud.NewClient(
		hcloud.WithHetznerEndpoint(server.URL),
		hcloud.WithToken("test-token"),
		hcloud.WithRetryOpts(hcloud.RetryOpts{MaxRetries: 0}),
	)
}

// storageBoxSchema builds an API payload for a single Storage Box.
func storageBoxSchema(id int64, name string, status hcloud.StorageBoxStatus, stats schema.StorageBoxStats, typeSize int64) schema.StorageBox {
	return schema.StorageBox{
		ID:     id,
		Name:   name,
		Status: string(status),
		StorageBoxType: schema.StorageBoxType{
			ID:   1,
			Name: "bx11",
			Size: typeSize,
		},
		Stats: stats,
	}
}

// listResponse mirrors the Storage Box list payload. The generated schema type
// carries no pagination meta, which the client needs to iterate pages.
type listResponse struct {
	StorageBoxes []schema.StorageBox `json:"storage_boxes"`
	Meta         schema.Meta         `json:"meta"`
}

// listHandler serves the Storage Box list endpoint, returning one page per
// entry in pages.
func listHandler(t *testing.T, pages ...[]schema.StorageBox) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/storage_boxes") {
			t.Errorf("unexpected request path %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		page := 1
		if raw := r.URL.Query().Get("page"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				t.Errorf("unparseable page parameter %q: %v", raw, err)
			}
			page = parsed
		}
		if page < 1 || page > len(pages) {
			t.Errorf("requested page %d out of range 1..%d", page, len(pages))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp := listResponse{
			StorageBoxes: pages[page-1],
			Meta: schema.Meta{Pagination: &schema.MetaPagination{
				Page:         page,
				PerPage:      len(pages[page-1]),
				LastPage:     len(pages),
				TotalEntries: totalEntries(pages),
			}},
		}
		if page < len(pages) {
			resp.Meta.Pagination.NextPage = page + 1
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	})
}

func totalEntries(pages [][]schema.StorageBox) int {
	total := 0
	for _, page := range pages {
		total += len(page)
	}
	return total
}

// resetMetrics clears the package level gauges, which are shared between tests.
func resetMetrics(t *testing.T) {
	t.Helper()

	for _, metric := range []*prometheus.GaugeVec{status, size, sizeData, sizeSnapshots, typeSize} {
		metric.Reset()
	}
}

func TestScrapeMetrics(t *testing.T) {
	resetMetrics(t)

	client := newTestClient(t, listHandler(t, []schema.StorageBox{
		storageBoxSchema(1, "box-a", hcloud.StorageBoxStatusActive, schema.StorageBoxStats{
			Size:          1073741824,
			SizeData:      536870912,
			SizeSnapshots: 268435456,
		}, 1099511627776),
		storageBoxSchema(2, "box-b", hcloud.StorageBoxStatusLocked, schema.StorageBoxStats{
			Size:          10,
			SizeData:      6,
			SizeSnapshots: 4,
		}, 2199023255552),
	}))

	if err := scrapeMetrics(t.Context(), client); err != nil {
		t.Fatalf("scrapeMetrics() returned error: %v", err)
	}

	// Every box gets one series per known status, so probing below cannot
	// silently create a missing series.
	if got, want := testutil.CollectAndCount(status), 2*len(StorageBoxStatusList); got != want {
		t.Errorf("status series count = %d, want %d", got, want)
	}

	wantStatus := map[string]hcloud.StorageBoxStatus{
		"box-a": hcloud.StorageBoxStatusActive,
		"box-b": hcloud.StorageBoxStatusLocked,
	}
	for name, active := range wantStatus {
		for _, s := range StorageBoxStatusList {
			want := 0.0
			if s == active {
				want = 1.0
			}
			if got := testutil.ToFloat64(status.WithLabelValues(name, string(s))); got != want {
				t.Errorf("status{storage-box=%q,status=%q} = %v, want %v", name, s, got, want)
			}
		}
	}

	for _, tc := range []struct {
		name   string
		metric *prometheus.GaugeVec
		box    string
		want   float64
	}{
		{"size box-a", size, "box-a", 1073741824},
		{"size_data box-a", sizeData, "box-a", 536870912},
		{"size_snapshots box-a", sizeSnapshots, "box-a", 268435456},
		{"type_size box-a", typeSize, "box-a", 1099511627776},
		{"size box-b", size, "box-b", 10},
		{"size_data box-b", sizeData, "box-b", 6},
		{"size_snapshots box-b", sizeSnapshots, "box-b", 4},
		{"type_size box-b", typeSize, "box-b", 2199023255552},
	} {
		if got := testutil.ToFloat64(tc.metric.WithLabelValues(tc.box)); got != tc.want {
			t.Errorf("%s = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// A box that disappears between scrapes must not leave a stale series behind.
func TestScrapeMetricsDropsStaleSeries(t *testing.T) {
	resetMetrics(t)

	client := newTestClient(t, listHandler(t,
		[]schema.StorageBox{
			storageBoxSchema(1, "box-a", hcloud.StorageBoxStatusActive, schema.StorageBoxStats{Size: 1}, 100),
			storageBoxSchema(2, "box-b", hcloud.StorageBoxStatusActive, schema.StorageBoxStats{Size: 2}, 200),
		},
	))

	if err := scrapeMetrics(t.Context(), client); err != nil {
		t.Fatalf("first scrapeMetrics() returned error: %v", err)
	}
	if got, want := testutil.CollectAndCount(size), 2; got != want {
		t.Fatalf("size series count after first scrape = %d, want %d", got, want)
	}

	client = newTestClient(t, listHandler(t, []schema.StorageBox{
		storageBoxSchema(1, "box-a", hcloud.StorageBoxStatusActive, schema.StorageBoxStats{Size: 1}, 100),
	}))

	if err := scrapeMetrics(t.Context(), client); err != nil {
		t.Fatalf("second scrapeMetrics() returned error: %v", err)
	}

	if got, want := testutil.CollectAndCount(size), 1; got != want {
		t.Errorf("size series count after second scrape = %d, want %d", got, want)
	}
	if got, want := testutil.CollectAndCount(status), len(StorageBoxStatusList); got != want {
		t.Errorf("status series count after second scrape = %d, want %d", got, want)
	}
}

func TestScrapeMetricsPaginates(t *testing.T) {
	resetMetrics(t)

	client := newTestClient(t, listHandler(t,
		[]schema.StorageBox{
			storageBoxSchema(1, "box-a", hcloud.StorageBoxStatusActive, schema.StorageBoxStats{Size: 1}, 100),
		},
		[]schema.StorageBox{
			storageBoxSchema(2, "box-b", hcloud.StorageBoxStatusInitializing, schema.StorageBoxStats{Size: 2}, 200),
		},
	))

	if err := scrapeMetrics(t.Context(), client); err != nil {
		t.Fatalf("scrapeMetrics() returned error: %v", err)
	}

	if got, want := testutil.CollectAndCount(size), 2; got != want {
		t.Fatalf("size series count = %d, want %d", got, want)
	}
	if got := testutil.ToFloat64(size.WithLabelValues("box-b")); got != 2 {
		t.Errorf("size{storage-box=\"box-b\"} = %v, want 2", got)
	}
}

func TestScrapeMetricsNoStorageBoxes(t *testing.T) {
	resetMetrics(t)

	client := newTestClient(t, listHandler(t, []schema.StorageBox{}))

	if err := scrapeMetrics(t.Context(), client); err != nil {
		t.Fatalf("scrapeMetrics() returned error: %v", err)
	}

	for name, metric := range map[string]*prometheus.GaugeVec{
		"status": status, "size": size, "size_data": sizeData,
		"size_snapshots": sizeSnapshots, "type_size": typeSize,
	} {
		if got := testutil.CollectAndCount(metric); got != 0 {
			t.Errorf("%s series count = %d, want 0", name, got)
		}
	}
}

func TestScrapeMetricsAPIError(t *testing.T) {
	resetMetrics(t)

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":"service_error","message":"boom"}}`))
	}))

	err := scrapeMetrics(t.Context(), client)
	if err == nil {
		t.Fatal("scrapeMetrics() = nil, want error")
	}
	if !strings.Contains(err.Error(), "error fetching storage boxes") {
		t.Errorf("error = %q, want it to mention fetching storage boxes", err)
	}
}

// A failed scrape must leave the previously exported values untouched rather
// than publishing a half-populated or empty snapshot.
func TestScrapeMetricsKeepsPreviousValuesOnError(t *testing.T) {
	resetMetrics(t)

	client := newTestClient(t, listHandler(t, []schema.StorageBox{
		storageBoxSchema(1, "box-a", hcloud.StorageBoxStatusActive, schema.StorageBoxStats{Size: 7}, 100),
	}))
	if err := scrapeMetrics(t.Context(), client); err != nil {
		t.Fatalf("scrapeMetrics() returned error: %v", err)
	}

	failing := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	if err := scrapeMetrics(t.Context(), failing); err == nil {
		t.Fatal("scrapeMetrics() = nil, want error")
	}

	if got := testutil.ToFloat64(size.WithLabelValues("box-a")); got != 7 {
		t.Errorf("size{storage-box=\"box-a\"} = %v, want 7", got)
	}
}

func TestScrapeMetricsCanceledContext(t *testing.T) {
	resetMetrics(t)

	client := newTestClient(t, listHandler(t, []schema.StorageBox{
		storageBoxSchema(1, "box-a", hcloud.StorageBoxStatusActive, schema.StorageBoxStats{Size: 1}, 100),
	}))

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if err := scrapeMetrics(ctx, client); err == nil {
		t.Fatal("scrapeMetrics() = nil, want error for canceled context")
	}
}

func TestHealthz(t *testing.T) {
	var ready atomic.Bool
	mux := newMux(&ready)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// healthz reports liveness, so it is OK even before the first scrape.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Body.String(), "ok\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestReadyz(t *testing.T) {
	for _, tc := range []struct {
		name     string
		ready    bool
		wantCode int
		wantBody string
	}{
		{"before first scrape", false, http.StatusServiceUnavailable, "not ready\n"},
		{"after first scrape", true, http.StatusOK, "ok\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ready atomic.Bool
			ready.Store(tc.ready)
			mux := newMux(&ready)

			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantCode)
			}
			if got := rec.Body.String(); got != tc.wantBody {
				t.Errorf("body = %q, want %q", got, tc.wantBody)
			}
		})
	}
}

func TestMetricsEndpoint(t *testing.T) {
	resetMetrics(t)

	client := newTestClient(t, listHandler(t, []schema.StorageBox{
		storageBoxSchema(1, "box-a", hcloud.StorageBoxStatusActive, schema.StorageBoxStats{Size: 42}, 100),
	}))
	if err := scrapeMetrics(t.Context(), client); err != nil {
		t.Fatalf("scrapeMetrics() returned error: %v", err)
	}

	var ready atomic.Bool
	ready.Store(true)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	newMux(&ready).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	for _, want := range []string{"storage_box_status", "storage_box_stats_size", "storage_box_type_size", "box-a"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}
