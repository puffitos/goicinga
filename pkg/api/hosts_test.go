package api

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func Test_hosts_Get(t *testing.T) {
	tests := []struct {
		name     string
		hostName string
		want     *Host
		wantCode int
		wantBody string
		wantErr  bool
	}{
		{
			name:     "empty host name",
			hostName: "",
			want:     nil,
			wantCode: 0,
			wantBody: "",
			wantErr:  true,
		},
		{
			name:     "host not found",
			hostName: "test",
			want:     nil,
			wantCode: http.StatusNotFound,
			wantBody: `{"error":404,"status":"No objects found."}`,
			wantErr:  false,
		},
	}

	c := hosts{
		ic: newTestClient(),
	}

	httpmock.ActivateNonDefault(c.ic.Client)
	defer httpmock.DeactivateAndReset()

	for _, tt := range tests {

		httpmock.RegisterResponder("GET",
			fmt.Sprintf("%s/objects/hosts/%s", c.ic.Config.BaseURL, tt.hostName),
			httpmock.NewStringResponder(tt.wantCode, tt.wantBody))

		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Get(context.Background(), tt.hostName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}
