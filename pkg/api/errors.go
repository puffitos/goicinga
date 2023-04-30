package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Error is the error returned by the icinga API
type Error struct {
	// Err is the status code returned by the icinga API
	Err int `json:"error"`
	// Status is the status message returned by the icinga API
	Status string `json:"status"`
}

// Error returns a verbose error message
func (e *Error) Error() string {
	return fmt.Sprintf("Error code: %d, error msg: %s", e.Err, e.Status)
}

// WrapError wraps the error returned by the icinga API
func WrapError(errBody io.ReadCloser) error {
	apiError := &Error{}
	if err := json.NewDecoder(errBody).Decode(apiError); err != nil {
		return err
	}
	return apiError
}

// IsNotFound returns true if the given error is of type Error is a 404.
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Err == http.StatusNotFound
	}
	return false
}
