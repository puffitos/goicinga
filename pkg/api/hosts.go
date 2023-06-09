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

type Host struct {
	CheckableAttrs
	// A short description of the host (e.g. displayed by external interfaces instead of the name if set).
	DisplayName string `json:"display_name,omitempty"`
	// The host’s IPv4 address. Available as command runtime macro $address$ if set.
	Address string `json:"address,omitempty"`
	// The host’s IPv6 address. Available as command runtime macro $address6$ if set.
	Address6 string `json:"address_6,omitempty"`
	// A list of host groups this host belongs to.
	Groups        []string  `json:"groups,omitempty"`
	LastHardState int       `json:"last_hard_state,omitempty"`
	LastState     int       `json:"last_state,omitempty"`
	LastStateDown time.Time `json:"last_state_down,omitempty"`
	LastStateUp   time.Time `json:"last_state_up,omitempty"`
	State         HostState `json:"state,omitempty"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The data is expected to be the binary representation of a ObjectQueryResult.
func (h *Host) UnmarshalJSON(data []byte) error {
	var oqr ObjectQueryResult
	if err := json.Unmarshal(data, &oqr); err != nil {
		return err
	}

	// iterate over all fields of the Host struct and set them to their
	// corresponding fields found in the oqr.Attrs map.
	elem := reflect.ValueOf(h).Elem()
	fields := getStructFields(elem.Type())
	var found int
	for _, v := range fields {
		n := strings.Split(v.Tag.Get("json"), ",")[0]
		if oqrv, ok := oqr.Attrs[n]; ok {
			found++
			typ := v.Type
			// The State field is of type HostState, but the API returns an int.
			// This makes sure that the CheckResult.State field is of type HostState.
			if v.Name == "State" {
				typ = reflect.TypeOf(HostState(0))
			}
			if typ == reflect.TypeOf(time.Time{}) {
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

	type Alias Host
	return json.Unmarshal(data, (*Alias)(h))
}

func getStructFields(t reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	// Iterate over the fields of the struct.
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Structs must be handled recursively.
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			nestedFields := getStructFields(field.Type)
			fields = append(fields, nestedFields...)
		} else {
			fields = append(fields, field)
		}
	}

	return fields
}

type HostState int

const (
	HostUp HostState = iota
	HostDown
)

// Hosts is the interface for interacting with Icinga hosts.
type Hosts interface {
	Get(ctx context.Context, name string) (*Host, error)
	Create(ctx context.Context, host *Host) error
	Update(ctx context.Context, host *Host) error
	Delete(ctx context.Context, name string, cascade bool) error
}

// hosts implements the Hosts interface.
type hosts struct {
	ic *Icinga
}

// newHostsClient returns a new Hosts client.
func newHostsClient(cfg *Config, log *logr.Logger) *hosts {
	l := log.WithName("hosts")
	return &hosts{ic: New(cfg, &l)}
}

// Get returns the host with the given name, or nil if it doesn't exist.
func (c *hosts) Get(ctx context.Context, name string) (*Host, error) {
	if name == "" {
		return nil, &NoIdentifierError{Object: "host"}
	}
	var res Host
	err := c.ic.
		Get().
		Endpoint("objects").
		Type("hosts").
		Object(name).
		Call(ctx).
		Into(&res)
	return &res, err
}

func (c *hosts) Update(ctx context.Context, host *Host) error {
	if host == nil {
		return fmt.Errorf("host cannot be nil")
	}

	b := &UpdateObjectRequest[Host]{
		Attrs: *host,
	}

	res := c.ic.Post().
		Endpoint("objects").
		Type("hosts").
		Object(host.Name).
		Body(b).
		Call(ctx)
	return res.Error()
}

// Create creates the given host.
func (c *hosts) Create(ctx context.Context, host *Host) error {
	if host == nil {
		return fmt.Errorf("host cannot be nil")
	}

	b := &CreateObjectRequest[CheckableAttrs]{
		Templates: host.Templates,
		Attrs:     host.CheckableAttrs,
	}

	res := c.ic.Put().
		Endpoint("objects").
		Type("hosts").
		Object(host.Name).
		Body(b).
		Call(ctx)
	return res.Error()
}

// Delete deletes the host with the given name.
func (c *hosts) Delete(ctx context.Context, name string, cascade bool) error {
	if name == "" {
		return &NoIdentifierError{Object: "host"}
	}

	b := &deleteObjectRequest{
		Cascade: cascade,
	}

	res := c.ic.Delete().
		Endpoint("objects").
		Type("hosts").
		Object(name).
		Body(b).
		Call(ctx)
	return res.Error()
}
