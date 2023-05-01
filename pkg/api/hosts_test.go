package api

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

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
			wantErr:  true,
		},
		{
			name:     "success",
			hostName: "test",
			want:     testHost(),
			wantCode: http.StatusOK,
			wantBody: testHostQueryResult(),
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
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, wantFields %v", got, tt.want)
			}
		})
	}
}

// testHostQueryResult creates a QueryResult object for a host returned by the testHostfunction,
// and the marshals it to a json string.
func testHostQueryResult() string {
	h := testHost()
	qr := ObjectQueryResult{
		Name: h.Name,
		Type: h.Type,
		Attrs: map[string]interface{}{
			"address":      h.Address,
			"display_name": h.DisplayName,
			"state":        h.State,
		},
		Joins: nil,
		Meta:  nil,
	}

	b, _ := qr.MarshalJSON()

	return string(b)
}

func TestHost_UnmarshalJSON(t *testing.T) {
	type fields struct {
		CheckableAttrs CheckableAttrs
		DisplayName    string
		Address        string
		Address6       string
		Groups         []string
		LastHardState  int
		LastState      int
		LastStateDown  time.Time
		LastStateUp    time.Time
		State          HostState
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "empty data",
			args:    args{data: []byte("")},
			fields:  fields{},
			wantErr: true,
		},
		{
			name:    "invalid data",
			args:    args{data: []byte(`{"test": "test"}`)},
			fields:  fields{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Host{}
			if err := h.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			expected := &Host{
				CheckableAttrs: tt.fields.CheckableAttrs,
				DisplayName:    tt.fields.DisplayName,
				Address:        tt.fields.Address,
				Address6:       tt.fields.Address6,
				Groups:         tt.fields.Groups,
				LastHardState:  tt.fields.LastHardState,
				LastState:      tt.fields.LastState,
				LastStateDown:  tt.fields.LastStateDown,
				LastStateUp:    tt.fields.LastStateUp,
				State:          tt.fields.State,
			}
			if !reflect.DeepEqual(h, expected) {
				t.Errorf("UnmarshalJSON() got = %v, wantFields %v", h, expected)
			}
		})
	}
}

// testHost returns a test host object.
func testHost() *Host {
	return &Host{
		CheckableAttrs: CheckableAttrs{
			CustomVarAttrs: CustomVarAttrs{
				ConfigObjectAttrs: ConfigObjectAttrs{
					Name: "test-host",
					ObjectAttrs: ObjectAttrs{
						Type: "Host",
					},
				},
			},
		},
		DisplayName: "test-host",
		Address:     "localhost",
		State:       HostUp,
	}
}

func Test_getStructFields(t *testing.T) {
	type args struct {
		t reflect.Type
	}

	// test structs
	type embed struct {
		EmbeddedTest string
	}
	type doubleEmbed struct {
		embed
		DoubleEmbeddedTest string
	}
	type structWithEmbed struct {
		Test string
		embed
	}
	type structWithDoubleEmbed struct {
		Test string
		doubleEmbed
	}

	tests := []struct {
		name       string
		args       args
		wantFields []string
	}{
		{
			name:       "empty",
			args:       args{t: reflect.TypeOf(struct{}{})},
			wantFields: nil,
		},
		{
			name: "struct with embed",
			args: args{t: reflect.TypeOf(structWithEmbed{})},
			wantFields: []string{
				"Test",
				"EmbeddedTest",
			},
		},
		{
			name: "double embed",
			args: args{t: reflect.TypeOf(structWithDoubleEmbed{})},
			wantFields: []string{
				"Test",
				"EmbeddedTest",
				"DoubleEmbeddedTest",
			},
		},
		{
			name: "host",
			args: args{t: reflect.TypeOf(ConfigObjectAttrs{})},
			wantFields: []string{
				"Type",
				"Name",
				"Active",
				"Extensions",
				"HAMode",
				"OriginalAttributes",
				"Package",
				"PauseCalled",
				"Paused",
				"ResumeCalled",
				"SourceLocation",
				"StartCalled",
				"StateLoaded",
				"StopCalled",
				"Templates",
				"Version",
				"Zone",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStructFields(tt.args.t)
			var gotFields []string
			for _, f := range got {
				gotFields = append(gotFields, f.Name)
			}

			if !reflect.DeepEqual(gotFields, tt.wantFields) {
				t.Errorf("getStructFields() = %v, wantFields %v", gotFields, tt.wantFields)
			}
		})
	}
}
