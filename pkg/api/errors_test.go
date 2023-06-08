package api

import (
	"testing"
)

func Test_WrapError_Single(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectedErr error
	}{
		{
			name: "valid single error",
			data: `{"error":404,"status":"No objects found."}`,
			expectedErr: &IcingaError{
				Err:    404,
				Status: "No objects found.",
			},
		},
	}

	for _, tt := range tests {

		data := []byte(tt.data)
		t.Run(tt.name, func(t *testing.T) {
			err := WrapError(data)
			if err == nil && tt.expectedErr != nil {
				t.Fatalf("expected error, got nil")
			}

			if e, ok := err.(*IcingaError); ok {
				if e.Err != tt.expectedErr.(*IcingaError).Err {
					t.Fatalf("expected error code %d, got %d", tt.expectedErr.(*IcingaError).Err, e.Err)
				}
			}
		})
	}
}

func Test_WrapError_Multiple(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectedErr error
	}{
		{
			name: "results array with an internal error",
			data: `{"results":[{"code":500,"name":"icinga-master","status":"Attribute 'statis' could not be set: Error: Invalid field ID.","type":"Host"}]}`,
			expectedErr: &IcingaInternalError{
				Code:   500,
				Name:   "icinga-master",
				Status: "Attribute 'statis' could not be set: Error: Invalid field ID.",
				Type:   "Host",
			},
		},
	}

	for _, tt := range tests {

		data := []byte(tt.data)
		t.Run(tt.name, func(t *testing.T) {
			err := WrapError(data)
			if err == nil && tt.expectedErr != nil {
				t.Fatalf("expected error, got nil")
			}

			if e, ok := err.(*IcingaError); ok {
				if e.Err != tt.expectedErr.(*IcingaError).Err {
					t.Fatalf("expected error code %d, got %d", tt.expectedErr.(*IcingaError).Err, e.Err)
				}
			}
		})
	}
}
