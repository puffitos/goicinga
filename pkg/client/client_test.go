package client

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/jarcoal/httpmock"
	"go.uber.org/zap"
)

// TestClient_Call tests the Call method of the client.
// It tests whether the headers are set correctly and
// whether the request is executed correctly.
func TestClient_Call(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
		wantBody string
		wantErr  bool
	}{
		{
			name:     "success",
			wantCode: http.StatusOK,
			wantBody: "OK",
			wantErr:  false,
		},
		{
			name:     "error creating request",
			wantCode: 0,
			wantBody: "",
			wantErr:  true,
		},
		{
			name:     "non-200 response",
			wantCode: http.StatusNotFound,
			wantBody: "Not Found",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient()
			req, err := http.NewRequestWithContext(context.Background(),
				"GET", "https://localhost:5665/", nil)
			if err != nil {
				t.Errorf("Call() failed to create request: %v", err)
				return
			}

			// httpmock
			httpmock.ActivateNonDefault(c.Conn)
			defer httpmock.DeactivateAndReset()
			if !tt.wantErr {
				httpmock.RegisterResponder("GET", "https://localhost:5665/", httpmock.NewStringResponder(tt.wantCode, tt.wantBody))
			} else {
				httpmock.RegisterResponder("GET", "https://localhost:5665/", httpmock.NewErrorResponder(fmt.Errorf("error creating request")))
			}

			got, _, err := c.Call(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Call() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && got != nil {
				t.Errorf("Call() got = %v, want nil", got)
			}
			if !tt.wantErr && !reflect.DeepEqual(got.StatusCode, tt.wantCode) {
				t.Errorf("Call() got = %v, wantCode %v", got, tt.wantCode)
			}
			if req.Header.Get("Accept") != "application/json" {
				t.Errorf("Call() failed to set Accept header")
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Call() failed to set Content-Type header")
			}
			if req.Header.Get("X-HTTP-Method-Override") != "GET" {
				t.Errorf("Call() failed to set X-HTTP-Method-Override header")
			}
		})
	}
}

// newTestClient returns a new client for testing purposes.
func newTestClient() *Client {
	var log logr.Logger

	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)

	return NewClient(&Config{
		BaseURL:  "https://localhost:5665",
		APIUser:  "test",
		APIPass:  "test",
		Timeout:  10 * time.Second,
		CertPath: "",
	}, &log)
}
