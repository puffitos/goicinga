package api

import (
	"context"
	"fmt"
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
