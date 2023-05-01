package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Request holds the request to the icinga API
type Request struct {
	c *Icinga

	// GET, POST, PUT, DELETE
	verb string
	// objects, actions, events, config, types, variables, status, templates
	endpoint string
	// required for config objects, i.e Hosts & Services
	typ string
	// the object's name or action to be performed
	object string

	body io.Reader
	err  error
}

// NewRequest creates a new request for the given verb
func NewRequest(c *Icinga) *Request {
	return &Request{
		c: c,
	}
}

// Endpoint sets the endpoint for the request
func (r *Request) Endpoint(endpoint string) *Request {
	if !allowedEndpoints(endpoint) {
		r.err = fmt.Errorf("invalid endpoint %s", endpoint)
		return r
	}
	r.endpoint = endpoint
	return r
}

// Type sets the type for the request
func (r *Request) Type(typ string) *Request {
	if !allowedType(typ) {
		r.err = fmt.Errorf("invalid type %s", typ)
		return r
	}
	if r.endpoint != "objects" {
		r.err = fmt.Errorf("type is only valid for endpoint objects")
		return r
	}
	if typ == "" {
		r.err = fmt.Errorf("type must not be empty")
		return r
	}
	r.typ = typ
	return r
}

// Object sets the object for the request
func (r *Request) Object(object string) *Request {
	if object == "" {
		r.err = fmt.Errorf("object must not be empty")
		return r
	}
	r.object = object
	return r
}

// Body sets the body for the request
func (r *Request) Body(body interface{}) *Request {
	if r.err != nil {
		return r
	}
	switch t := body.(type) {
	case io.Reader:
		r.body = t
	case []byte:
		r.body = bytes.NewReader(t)
	default:
		b, err := json.Marshal(body)
		if err != nil {
			r.err = err
			return r
		}
		r.body = bytes.NewReader(b)
	}
	return r
}

func allowedEndpoints(endpoint string) bool {
	switch endpoint {
	case "objects", "actions", "events", "config", "types", "variables", "status", "templates":
		return true
	default:
		return false
	}
}

func allowedType(typ string) bool {
	allowedTypes := []string{"hosts", "services"}
	for _, t := range allowedTypes {
		if t == typ {
			return true
		}
	}
	return false
}

// Call executes the given request and returns the result of that call.
// Returns an error in the Result if the call failed. http.Client errors are returned directly,
// icinga API errors are wrapped in an api.IcingaError.
func (r *Request) Call(ctx context.Context) *Result {
	var res Result
	if r.err != nil {
		res.err = r.err
		return &res
	}
	r.c.Log.V(1).Info("calling icinga api", "endpoint", r.endpoint, "object", r.object, "method", r.verb, "body", r.body)
	req, err := http.NewRequestWithContext(
		ctx,
		r.verb,
		strings.Join([]string{r.c.Config.BaseURL, r.endpoint, r.typ, r.object}, "/"),
		r.body,
	)
	if err != nil {
		res.err = err
		r.c.Log.Error(err, "failed creating request",
			"endpoint", r.endpoint,
			"object", r.object,
			"method", r.verb,
			"body", r.body)
		return &res
	}
	req.SetBasicAuth(r.c.Config.APIUser, r.c.Config.APIPass)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-HTTP-Method-Override", req.Method)

	resp, err := r.c.Client.Do(req) //nolint:bodyclose
	if err != nil {
		res.err = err
		r.c.Log.Error(err, "failed to Call icinga api", "endpoint", req.URL.Path)
		return &res
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			r.c.Log.Error(err, "failed closing response body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		res.err = err
		r.c.Log.Error(err, "failed reading response body")
		return &res
	}
	r.c.Log.V(1).Info("response from icinga api", "status", resp.StatusCode, "body", string(body))
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		res.err = WrapError(body)
	}
	res.statusCode = resp.StatusCode
	res.body = body
	return &res
}

// Result holds the response from the icinga API
type Result struct {
	statusCode int
	body       []byte
	err        error
}

// Into decodes the response body into the given interface
func (r *Result) Into(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	return json.Unmarshal(r.body, v)
}

// Error returns any error that occurred during the actual call
// or nil if the call was successful
func (r *Result) Error() error {
	return r.err
}
