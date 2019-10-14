package metrics

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	next := m.next

	start := time.Now()

	// Record response to get status code and size of the reply.
	rw := httpserver.NewResponseRecorder(w)
	// Get time to first write.
	tw := &timedResponseWriter{ResponseWriter: rw}

	status, err := next.ServeHTTP(tw, r)

	// If nothing was explicitly written, consider the request written to
	// now that it has completed.
	tw.didWrite()

	// Transparently capture the status code so as to not side effect other plugins
	stat := status
	if err != nil && status == 0 {
		// Some middlewares set the status to 0, but return an non nil error: map these to status 500
		stat = 500
	} else if status == 0 {
		// 'proxy' returns a status code of 0, but the actual status is available on rw.
		// Note that if 'proxy' encounters an error, it returns the appropriate status code (such as 502)
		// from ServeHTTP and is captured above with 'stat := status'.
		stat = rw.Status()
	}

	path := "/-"
	if stat != 404 {
		path = getPath(m, r.URL.String())
	}

	// We only want 2xx, 3xx, 4xx, 5xx
	statusStr := string(strconv.Itoa(stat)[0]) + "xx"

	host, _, _ := net.SplitHostPort(r.Host)
	host = strings.ToLower(host)

	requestCount.WithLabelValues(host, path).Inc()
	requestDuration.WithLabelValues(host, path).Observe(time.Since(start).Seconds())
	responseSize.WithLabelValues(host, path, statusStr).Observe(float64(rw.Size()))
	responseStatus.WithLabelValues(host, path, statusStr).Inc()
	responseLatency.WithLabelValues(host, path, statusStr).Observe(tw.firstWrite.Sub(start).Seconds())

	return status, err
}

// A timedResponseWriter tracks the time when the first response write
// happened.
type timedResponseWriter struct {
	firstWrite time.Time
	http.ResponseWriter
}

func (w *timedResponseWriter) didWrite() {
	if w.firstWrite.IsZero() {
		w.firstWrite = time.Now()
	}
}

func (w *timedResponseWriter) Write(data []byte) (int, error) {
	w.didWrite()
	return w.ResponseWriter.Write(data)
}

func (w *timedResponseWriter) WriteHeader(statuscode int) {
	// We consider this a write as it's valid to respond to a request by
	// just setting a status code and returning.
	w.didWrite()
	w.ResponseWriter.WriteHeader(statuscode)
}

func getPath(m *Metrics, url string) string {
	re := m.compiledRegex
	submatch := re.FindSubmatch([]byte(url))

	if len(submatch) == 2 {
		return string(submatch[1])
	}

	return "/-"
}
