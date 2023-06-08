package api

import (
	"encoding/json"
	"fmt"
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

type NoIdentifierError struct {
	// The type of the object
	Object string
}

func (e *NoIdentifierError) Error() string {
	return fmt.Sprintf("no identifier provided for object %s", e.Object)
}
