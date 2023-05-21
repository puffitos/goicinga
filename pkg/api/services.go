package api

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

type ServiceState int

const (
	ServiceOk ServiceState = iota
	ServiceWarning
	ServiceCritical
	ServiceUnknown
)

type Service struct {
	CheckableAttrs
	DisplayName       string       `json:"display_name"`
	Groups            []string     `json:"groups"`
	HostName          string       `json:"host_name"`
	LastHardState     int          `json:"last_hard_state"`
	LastState         int          `json:"last_state"`
	LastStateCritical time.Time    `json:"last_state_critical"`
	LastStateOK       time.Time    `json:"last_state_ok"`
	LastStateUnknown  time.Time    `json:"last_state_unknown"`
	LastStateWarning  time.Time    `json:"last_state_warning"`
	State             ServiceState `json:"state"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The data is expected to be the binary representation of a ObjectQueryResult.
func (s *Service) UnmarshalJSON(data []byte) error {
	var oqr ObjectQueryResult
	if err := json.Unmarshal(data, &oqr); err != nil {
		return err
	}

	// iterate over all fields of the Service struct and set them to their
	// corresponding fields found in the oqr.Attrs map.
	elem := reflect.ValueOf(s).Elem()
	fields := getStructFields(elem.Type())
	var found int
	for _, v := range fields {
		n := strings.Split(v.Tag.Get("json"), ",")[0]
		if oqrv, ok := oqr.Attrs[n]; ok {
			found++
			typ := v.Type
			// The State field is of type ServiceState, but the API returns an int.
			// This makes sure that the CheckResult.State field is of type ServiceState.
			switch v.Name {
			case "State":
				typ = reflect.TypeOf(ServiceState(0))
			case "Groups":
				// iterate over the slice and convert each element to the correct type.
				slice := reflect.MakeSlice(typ, 0, 0)
				for _, e := range oqrv.([]interface{}) {
					slice = reflect.Append(slice, reflect.ValueOf(e))
				}
				elem.FieldByName(v.Name).Set(slice)
				continue
			}
			if typ == reflect.TypeOf(time.Time{}) {
				// Convert the time string to a time.Time object.
				// The time string is a unix timestamp in the format
				// seconds.nanoseconds
				seconds := int64(oqrv.(float64))

				nanoseconds := int64((oqrv.(float64) - float64(seconds)) * Ms)
				t := time.Unix(seconds, nanoseconds).UTC()
				elem.FieldByName(v.Name).Set(reflect.ValueOf(t))
				continue
			}
			value := reflect.ValueOf(oqrv).Convert(typ)
			elem.FieldByName(v.Name).Set(value)
		}
	}

	if found == 0 {
		return fmt.Errorf("no known fields found in Attrs map of ObjectQueryResult")
	}

	type Alias Service
	return json.Unmarshal(data, (*Alias)(s))
}

type Services interface {
	Get(ctx context.Context, name string) (*Service, error)
	Create(ctx context.Context, svc *Service) error
	Delete(ctx context.Context, name string, cascade bool) error
}

// services implements the Services interface.
type services struct {
	ic *Icinga
}

// newServicesClient returns a new Services client.
func newServicesClient(cfg *Config, log *logr.Logger) *services {
	l := log.WithName("services")
	return &services{ic: New(cfg, &l)}
}

// Get returns the service with the given name on the given host.
func (c *services) Get(ctx context.Context, name string) (*Service, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	var res Service
	err := c.ic.Get().
		Endpoint("objects").
		Type("services").
		Object(name).
		Call(ctx).
		Into(&res)
	return &res, err
}

// Create creates a new types in Icinga with the given name, if it doesn't already exist.
func (c *services) Create(ctx context.Context, svc *Service) error {
	if svc == nil {
		return fmt.Errorf("service cannot be nil")
	}

	b := &CreateObjectRequest[CheckableAttrs]{
		Templates: svc.Templates,
		Attrs:     svc.CheckableAttrs,
	}
	res := c.ic.Put().
		Endpoint("objects").
		Type("services").
		Object(svc.Name).
		Body(b).
		Call(ctx)

	return res.Error()
}

// Delete deletes the given service from Icinga.
func (c *services) Delete(ctx context.Context, name string, cascade bool) error {
	type deleteServiceRequest struct {
		Cascade bool `json:"cascade"`
	}
	b := &deleteServiceRequest{Cascade: cascade}

	res := c.ic.Delete().
		Endpoint("objects").
		Type("services").
		Object(name).
		Body(b).
		Call(ctx)

	return res.Error()
}
