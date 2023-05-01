package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// IcingaError is the error returned by the icinga API
type IcingaError struct {
	// Err is the status code returned by the icinga API
	Err int `json:"error"`
	// Status is the status message returned by the icinga API
	Status string `json:"status"`
}

// IcingaError returns a verbose error message
func (e *IcingaError) Error() string {
	return fmt.Sprintf("IcingaError code: %d, error msg: %s", e.Err, e.Status)
}

// WrapError wraps the error returned by the icinga API
func WrapError(body []byte) error {
	var apiError IcingaError
	err := json.Unmarshal(body, &apiError)
	if err != nil {
		return err
	}
	return &apiError
}

// IsNotFound returns true if the given error is of type IcingaError is a 404.
func IsNotFound(err error) bool {
	if e, ok := err.(*IcingaError); ok {
		return e.Err == http.StatusNotFound
	}
	return false
}
