package api

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/go-logr/zapr"
	"github.com/jarcoal/httpmock"
	"github.com/kr/pretty"
	"go.uber.org/zap"
)

func Test_services_Get(t *testing.T) {
	tests := []struct {
		name     string
		svcName  string
		want     *Service
		wantCode int
		wantBody string
		wantErr  bool
	}{
		{
			name:     "empty service name",
			svcName:  "",
			want:     nil,
			wantCode: 0,
			wantBody: "",
			wantErr:  true,
		},
		{
			name:     "service not found",
			svcName:  "test",
			want:     nil,
			wantCode: http.StatusNotFound,
			wantBody: `{"error":404,"status":"No objects found."}`,
			wantErr:  true,
		},
		{
			name:     "success",
			svcName:  "test",
			want:     testService(),
			wantCode: http.StatusOK,
			wantBody: testServiceQueryResult(),
			wantErr:  false,
		},
	}

	c := services{
		ic: newTestClient(),
	}

	httpmock.ActivateNonDefault(c.ic.Client)
	defer httpmock.DeactivateAndReset()

	for _, tt := range tests {

		httpmock.RegisterResponder("GET",
			fmt.Sprintf("%s/objects/services/%s", c.ic.Config.BaseURL, tt.svcName),
			httpmock.NewStringResponder(tt.wantCode, tt.wantBody))

		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Get(context.Background(), tt.svcName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
				for _, diff := range pretty.Diff(got, tt.want) {
					t.Log(diff)
				}
			}
		})
	}
}

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

func TestService_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name     string
		expected *Service
		args     args
		wantErr  bool
	}{
		{
			name:     "empty data",
			args:     args{data: []byte("")},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid data",
			args:     args{data: []byte(`{"test": "test"}`)},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "success",
			args:     args{data: []byte(testServiceQueryResult())},
			expected: testService(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{}
			if tt.expected != nil {
				s = tt.expected
			}
			if err := s.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// testServiceQueryResult returns the expected response body
// for a successful service query.
func testServiceQueryResult() string {
	s := testService()
	// convert time.Time to unix timestamp with nanoseconds
	lastOK := float64(s.LastStateOK.UnixNano() / int64(time.Second))
	lastWarning := float64(s.LastStateWarning.UnixNano() / int64(time.Second))
	lastCritical := float64(s.LastStateCritical.UnixNano() / int64(time.Second))
	lastUnknown := float64(s.LastStateUnknown.UnixNano() / int64(time.Second))
	lastCheck := float64(s.LastCheck.UnixNano() / int64(time.Second))

	qr := ObjectQueryResult{
		Name: s.Name,
		Type: s.Type,
		Attrs: map[string]interface{}{
			"display_name":        s.DisplayName,
			"state":               s.State,
			"groups":              s.Groups,
			"check_command":       s.CheckCommand,
			"state_type":          s.StateType,
			"last_check":          lastCheck,
			"host_name":           s.HostName,
			"last_hard_state":     s.LastHardState,
			"last_state":          s.LastState,
			"last_state_ok":       lastOK,
			"last_state_critical": lastCritical,
			"last_state_unknown":  lastUnknown,
			"last_state_warning":  lastWarning,
			"severity":            s.Severity,
			"problem":             s.Problem,
		},
		Joins: nil,
		Meta:  nil,
	}

	b, _ := qr.MarshalJSON()
	return string(b)
}

// testService returns a new Service for testing.
func testService() *Service {
	lastCheck := time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)
	lastOK := time.Date(2020, 3, 1, 0, 0, 3, 0, time.UTC)
	lastCritical := time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)
	lastUnknown := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	lastWarning := time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC)
	return &Service{
		CheckableAttrs: CheckableAttrs{
			CustomVarAttrs: CustomVarAttrs{
				ConfigObjectAttrs: ConfigObjectAttrs{
					Name: "test-host!test-service",
				},
			},
			CheckCommand: "test",
			Problem:      false,
			Severity:     0,
			StateType:    StateTypeSoft,
			LastCheck:    lastCheck,
		},
		DisplayName:       "test-service",
		Groups:            []string{},
		HostName:          "test-host",
		LastHardState:     0,
		LastState:         0,
		LastStateCritical: lastCritical,
		LastStateOK:       lastOK,
		LastStateUnknown:  lastUnknown,
		LastStateWarning:  lastWarning,
		State:             0,
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
