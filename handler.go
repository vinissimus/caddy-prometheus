package metrics

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/caddy/middleware"
)

func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	next := m.next
	host, err := host(r)
	if err != nil {
		host = "-"
	}
	start := time.Now()

	// Record response to get status code and size of the reply.
	rw := middleware.NewResponseRecorder(w)
	status, err := next.ServeHTTP(rw, r)

	requestCount.WithLabelValues(host).Inc()
	requestDuration.WithLabelValues(host).Observe(float64(time.Since(start)) / float64(time.Second))
	responseSize.WithLabelValues(host).Observe(float64(rw.Size()))
	responseStatus.WithLabelValues(host, strconv.Itoa(rw.Status())).Inc()

	return status, err
}

func host(r *http.Request) (string, error) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		if !strings.Contains(r.Host, ":") {
			return r.Host, nil
		}
		return "", err
	}
	return strings.ToLower(host), nil
}
