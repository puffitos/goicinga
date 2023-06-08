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

// IcingaInternalError is the error returned by the icinga API when an operation on an object fails
type IcingaInternalError struct {
	// Code is the status code returned by the icinga API
	Code int `json:"code"`
	// Name is the name of the object that caused the error
	Name string `json:"name"`
	// Status is the status message returned by the icinga API
	Status string `json:"status"`
	// Type is the type of the object that caused the error
	Type string `json:"type"`
}

// Error returns the error message
func (e *IcingaInternalError) Error() string {
	return fmt.Sprintf("icinga internal error code: %d, name: %s, status: %s, type: %s", e.Code, e.Name, e.Status, e.Type)
}

// IcingaError returns a verbose error message
func (e *IcingaError) Error() string {
	return fmt.Sprintf("IcingaError code: %d, error msg: %s", e.Err, e.Status)
}

// WrapError wraps the error returned by the icinga API
func WrapError(body []byte) error {
	var d map[string]interface{}
	err := json.Unmarshal(body, &d)
	if err != nil {
		return err
	}

	// case 1: icinga internal error
	if _, ok := d["results"]; ok {
		results := d["results"].([]interface{})
		if len(results) == 0 || len(results) > 1 {
			return fmt.Errorf("invalid number of results in error response")
		}

		err := results[0].(map[string]interface{})
		// sanity check: check that all expected fields are present
		if _, ok := err["code"]; !ok {
			return fmt.Errorf("invalid error response")
		}
		if _, ok := err["name"]; !ok {
			return fmt.Errorf("invalid error response")
		}
		if _, ok := err["status"]; !ok {
			return fmt.Errorf("invalid error response")
		}
		if _, ok := err["type"]; !ok {
			return fmt.Errorf("invalid error response")
		}
		return &IcingaInternalError{
			Code:   int(err["code"].(float64)),
			Name:   err["name"].(string),
			Status: err["status"].(string),
			Type:   err["type"].(string),
		}
	}

	// case 2: icinga error
	var result IcingaError
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	return &result
}

type NoIdentifierError struct {
	// The type of the object
	Object string
}

func (e *NoIdentifierError) Error() string {
	return fmt.Sprintf("no identifier provided for object %s", e.Object)
}
