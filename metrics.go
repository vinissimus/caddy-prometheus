package prommetrics

import "github.com/prometheus/client_golang/prometheus"

const namespace = "caddy"

var (
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
	responseStatus  *prometheus.CounterVec
	responseLatency *prometheus.HistogramVec
)

func define(subsystem string) {
	if subsystem == "" {
		subsystem = "http"
	}
	requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "regex_request_count_total",
		Help:      "Counter of HTTP(S) requests made.",
	}, []string{"host", "path"})

	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "regex_request_duration_seconds",
		Help:      "Histogram of the time (in seconds) each request took.",
		Buckets:   append(prometheus.DefBuckets, 30, 60),
	}, []string{"host", "path", "status"})

	responseSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "regex_response_size_bytes",
		Help:      "Size of the returns response in bytes.",
		Buckets:   []float64{0, 1e3, 1e4, 1e5, 1e6, 5e6},
	}, []string{"host", "path", "status"})

	responseStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "regex_response_status_count_total",
		Help:      "Counter of response status codes.",
	}, []string{"host", "path", "status"})

	responseLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "regex_response_latency_seconds",
		Help:      "Histogram of the time (in seconds) until the first write for each request.",
		Buckets:   append(prometheus.DefBuckets, 30, 60),
	}, []string{"host", "path", "status"})
}
