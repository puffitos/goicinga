package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/zapr"
	"github.com/jarcoal/httpmock"
	"go.uber.org/zap"
)

func Test_services_Create(t *testing.T) {
	tests := []struct {
		name string
		svc  *Service

		mockBody string
		mockCode int
		wantErr  bool
	}{
		{
			name:     "nil service",
			svc:      nil,
			mockBody: "",
			mockCode: 0,
			wantErr:  true,
		},
		{
			name:     "success",
			svc:      testService(),
			mockBody: `{"results":[{"code":200.0,"name":"test","status":"Successfully created object 'test' of type 'Service'."}]}`,
			mockCode: http.StatusOK,
			wantErr:  false,
		},
		{
			name:     "service already exists",
			svc:      testService(),
			mockBody: `{"error":409,"status":"Object 'test' of type 'Service' already exists."}`,
			mockCode: http.StatusConflict,
			wantErr:  true,
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
			httpmock.RegisterResponder("PUT", fmt.Sprintf("https://icinga-server:5665/v1/objects/services/%s", tt.svc.Name),
				httpmock.NewStringResponder(tt.mockCode, tt.mockBody))

			if err := c.Create(context.Background(), tt.svc); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_services_Delete(t *testing.T) {
	type args struct {
		name    string
		cascade bool
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantBody string
		wantErr  bool
	}{
		{
			name: "success",
			args: args{
				name:    "test-host!test-service",
				cascade: false,
			},
			wantCode: http.StatusOK,
			wantBody: `{"results": [{"code": 200.0, "status": "Object was deleted successfully."}]}`,
			wantErr:  false,
		},
		{
			name: "non OK status code",
			args: args{
				name:    "test-host!test-service",
				cascade: false,
			},
			wantCode: http.StatusNotFound,
			wantBody: `{"error": "Object not found."}`,
			wantErr:  true,
		},
		{
			name: "fail reading body",
			args: args{
				name:    "test-host!test-service",
				cascade: false,
			},
			wantCode: 0,
			wantBody: "",
			wantErr:  true,
		},
	}

	c := &services{
		ic: newTestClient(),
	}

	httpmock.ActivateNonDefault(c.ic.Client)
	defer httpmock.DeactivateAndReset()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder("DELETE",
				fmt.Sprintf("https://icinga-server:5665/v1/objects/services/%s", tt.args.name),
				httpmock.NewStringResponder(tt.wantCode, tt.wantBody))

			if tt.name == "fail reading body" {
				httpmock.NewErrorResponder(fmt.Errorf("client error"))
			}

			if err := c.Delete(context.Background(), tt.args.name, tt.args.cascade); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// testService returns a new Service for testing.
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
func newTestClient() *Icinga {
	cfg := &Config{
		BaseURL: "https://icinga-server:5665/v1",
		APIUser: "root",
		APIPass: "root",
		Timeout: 1 * time.Second,
	}

	l := zapr.NewLogger(zap.NewExample().Named("test"))
	return New(cfg, &l)
}
