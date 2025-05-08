// metrics

package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of HTTP-request per month",
		},
		[]string{"url_path", "year_month"},
	)

	ResponseTimeHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Histogram of respone times for HTTP request",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func InitMetrics() {
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(ResponseTimeHistogram)
}

// CleanupMetrics unregisters all metrics to avoid memory leaks
// and ensures proper cleanup during application shutdown
func CleanupMetrics() {
	prometheus.Unregister(RequestCount)
	prometheus.Unregister(ResponseTimeHistogram)
}

func StartMetricServer(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf(`Error starting metrics serverer: %v\n`, err)
	}
}
