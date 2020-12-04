package prommetrics

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"net/http/httptest"

	"github.com/caddyserver/caddy/v2"
)

func next(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func TestMetrics_ServeHTTP(t *testing.T) {
	successRequest, err := http.NewRequest("GET", "http://test.com/success", nil)
	errorRequest, err := http.NewRequest("GET", "http://test.com/error", nil)
	proxyRequest, err := http.NewRequest("GET", "http://test.com/proxy", nil)
	// proxyErrorRequest, err := http.NewRequest("GET", "http://test.com/proxyerror", nil)

	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    observed
		wantErr bool
	}{
		{
			name: "200 handler response",
			args: args{
				w: httptest.NewRecorder(),
				r: successRequest,
			},
			want:    observed{status: "2xx", path: "/success"},
			wantErr: false,
		},
		{
			name: "500 handler response",
			args: args{
				w: httptest.NewRecorder(),
				r: errorRequest,
			},
			want:    observed{status: "5xx"},
			wantErr: false,
		},
		{
			name: "proxy handler response",
			args: args{
				w: httptest.NewRecorder(),
				r: proxyRequest,
			},
			want:    observed{status: "5xx"},
			wantErr: false,
		},
		// {
		// 	name: "proxy error handler response",
		// 	args: args{
		// 		w: httptest.NewRecorder(),
		// 		r: proxyErrorRequest,
		// 	},
		// 	want:    observed{},
		// 	wantErr: true,
		// },
	}

	o := &observed{}
	mockObserver := func(obs *observed) {
		o = obs
	}

	m := &Metrics{
		regex:    `^/([^/]*).*$`,
		observer: mockObserver,
	}

	// Compiles regex
	m.Provision(caddy.Context{})

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.ServeHTTP(tt.args.w, tt.args.r, testHandler{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Test %d: Metrics.ServeHTTP() error = %v, wantErr %v", i, err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(o, tt.want) {
				t.Errorf("Test %d: Metrics.ServeHTTP() = %+v, want %+v", i, o, tt.want)
			}
		})
	}
}

type testHandler struct{}

func (h testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	var err error

	switch r.URL.Path {
	case "/success":
		w.WriteHeader(200)
		w.Write([]byte{})
	case "/error":
		w.WriteHeader(500)
		w.Write([]byte{})
	case "/proxyerror":
		err = errors.New("no hosts available upstream")
	}

	return err
}
