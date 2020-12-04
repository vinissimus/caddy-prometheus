package prommetrics

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

type observed struct {
	start  time.Time
	host   string
	path   string
	status string
	ttfb   float64
	size   float64
}

func (m Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	obs := &observed{
		start: time.Now(),
	}
	// Based on https://github.com/caddyserver/caddy/blob/f197cec7f3a599ca18807e7b7719ef7666cfdb70/modules/caddyhttp/metrics.go#L121-L133
	var stat int

	// This is a _bit_ of a hack - it depends on the ShouldBufferFunc always
	// being called when the headers are written.
	// Effectively the same behaviour as promhttp.InstrumentHandlerTimeToWriteHeader.
	writeHeaderRecorder := caddyhttp.ShouldBufferFunc(func(status int, header http.Header) bool {
		stat = status
		obs.ttfb = time.Since(obs.start).Seconds()
		return false
	})
	wrec := caddyhttp.NewResponseRecorder(w, nil, writeHeaderRecorder)
	err := next.ServeHTTP(wrec, r)

	// Transparently capture the status code so as to not side effect other plugins
	if err != nil && stat == 0 {
		// Some middlewares set the status to 0, but return an non nil error: map these to status 500
		stat = 500
	}

	path := "/-"
	if stat != 404 {
		path = getPath(&m, r.URL.String())
	}

	host, _, _ := net.SplitHostPort(r.Host)
	host = strings.ToLower(host)

	obs.host = host
	obs.path = path
	// We only want 2xx, 3xx, 4xx, 5xx
	obs.status = string(sanitizeCode(stat)[0]) + "xx"
	obs.size = float64(wrec.Size())

	m.observer(obs)

	return nil
}

func observe(o *observed) {
	requestCount.WithLabelValues(o.host, o.path).Inc()
	requestDuration.WithLabelValues(o.host, o.path, o.status).Observe(time.Since(o.start).Seconds())
	responseSize.WithLabelValues(o.host, o.path, o.status).Observe(o.size)
	responseStatus.WithLabelValues(o.host, o.path, o.status).Inc()
	responseLatency.WithLabelValues(o.host, o.path, o.status).Observe(o.ttfb)
}

func getPath(m *Metrics, url string) string {
	re := m.compiledRegex
	submatch := re.FindSubmatch([]byte(url))

	if len(submatch) == 2 {
		return string(submatch[1])
	}

	return "/-"
}

func sanitizeCode(code int) string {
	if code == 0 {
		return "200"
	}
	return strconv.Itoa(code)
}
