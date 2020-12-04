package prommetrics

import (
	"regexp"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/prometheus/client_golang/prometheus"
)

// Init initializes the module
func init() {
	caddy.RegisterModule(Metrics{})
	httpcaddyfile.RegisterHandlerDirective("prometheus", parseCaddyfile)
}

const defaultRegex = `^/([^/]*).*$`

// Metrics holds the prometheus configuration.
type Metrics struct {
	regex string

	// subsystem?
	compiledRegex *regexp.Regexp

	observer func(*observed)
}

// CaddyModule returns the Caddy module information.
func (Metrics) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.prometheus",
		New: newModule,
	}
}

func newModule() caddy.Module {
	define("")

	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(responseLatency)
	prometheus.MustRegister(responseSize)
	prometheus.MustRegister(responseStatus)

	return Metrics{
		observer: observe,
	}
}

// Provision implements caddy.Provisioner.
func (m *Metrics) Provision(ctx caddy.Context) error {
	if len(m.regex) == 0 {
		m.regex = defaultRegex
	}
	m.compiledRegex = regexp.MustCompile(m.regex)
	return nil
}

// Validate implements caddy.Validator.
func (m Metrics) Validate() error {
	return nil
}

// UnmarshalCaddyfile expects the following syntax:
//
//	prometheus {
//		regex ^/([^/]*).*$
//	}
// Or just:
//
//	prometheus
//
func (m *Metrics) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		args := d.RemainingArgs()
		if len(args) > 0 {
			return d.Errf("prometheus: unexpected args: %v", args)
		}
		for d.NextBlock(0) {
			switch d.Val() {
			case "regex":
				if !d.Args(&m.regex) {
					return d.ArgErr()
				}
			default:
				return d.Errf("prometheus: unknown item: %s", d.Val())
			}
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Metrics
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}
