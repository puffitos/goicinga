package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/zapr"
	"github.com/jarcoal/httpmock"
	"github.com/puffitos/goicinga/pkg/client"
	"go.uber.org/zap"
)

func Test_services_Create(t *testing.T) {
	tests := []struct {
		name     string
		svc      *Service
		codeGet  int
		bodyGet  string
		codePost int
		wantErr  bool
	}{
		{
			name:     "nil service",
			svc:      nil,
			codeGet:  0,
			codePost: 0,
			wantErr:  true,
		},
		{
			name:     "success",
			svc:      testService(),
			codeGet:  http.StatusNotFound,
			bodyGet:  `{"error":"Object not found."}`,
			codePost: http.StatusOK,
			wantErr:  false,
		},
		{
			name:     "service already exists",
			svc:      testService(),
			codeGet:  http.StatusOK,
			bodyGet:  `{"attrs": {"name":"test"}}`,
			codePost: 0,
			wantErr:  false,
		},
	}

	c := &services{
		ic: newTestClient(),
	}

	httpmock.ActivateNonDefault(c.ic.Client)
	defer httpmock.DeactivateAndReset()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.svc == nil {
				tt.svc = &Service{}
			}
			httpmock.RegisterResponder("GET", fmt.Sprintf("https://localhost:5665/v1/objects/services/%s", tt.svc.Name),
				httpmock.NewStringResponder(tt.codeGet, tt.bodyGet))
			httpmock.RegisterResponder("PUT", fmt.Sprintf("https://localhost:5665/v1/objects/services/%s", tt.svc.Name),
				httpmock.NewStringResponder(tt.codePost, "does not matter"))

			if err := c.Create(context.Background(), tt.svc); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func testService() *Service {
	return &Service{
		CheckableAttrs: CheckableAttrs{
			CustomVarAttrs: CustomVarAttrs{
				ConfigObjectAttrs: ConfigObjectAttrs{
					Name: "test-host!test-service",
				},
			},
			CheckCommand: "test",
		},
	}
}

// newTestClient returns a new Icinga client for testing.
func newTestClient() *client.Icinga {
	cfg := &client.Config{
		BaseURL: "https://localhost:5665/v1",
		APIUser: "root",
		APIPass: "root",
		Timeout: 1 * time.Second,
	}

	l := zapr.NewLogger(zap.NewExample().Named("test"))
	return client.New(cfg, &l)
}
