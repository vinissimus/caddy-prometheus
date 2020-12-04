package prommetrics

import (
	"sync"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

var testInit = &sync.Once{}

func TestParse(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		expected  Metrics
	}{
		{`prometheus`, false, Metrics{regex: defaultRegex, init: &sync.Once{}}},
		{`prometheus {
			a b
		}`, true, Metrics{}},
		{`prometheus prometheus`, true, Metrics{init: &sync.Once{}}},
		{`prometheus {
			regex "^https?://([^\/]+).*$"
		}`, false, Metrics{regex: `^https?://([^\/]+).*$`, init: &sync.Once{}}},
	}

	for i, test := range tests {
		h := httpcaddyfile.Helper{
			Dispenser: caddyfile.NewTestDispenser(test.input),
		}
		actual, err := parseCaddyfile(h)
		got := actual.(Metrics)
		got.init = testInit
		got.Provision(caddy.Context{})

		if test.shouldErr {
			if err == nil {
				t.Errorf("Test %v: Expected error but found nil", i)
			}
		} else {
			if err != nil {
				t.Errorf("Test %v: Expected no error but found error: %v", i, err)
			} else if test.expected.regex != got.regex {
				t.Errorf("Test %v: Created Metrics (\n%#v\n) does not match expected (\n%#v\n)", i, got, test.expected)
			}
		}
	}
}
