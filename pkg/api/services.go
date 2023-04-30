package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/puffitos/goicinga/pkg/client"
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
	DisplayName       string    `json:"display_name"`
	Groups            []string  `json:"groups"`
	HostName          string    `json:"host_name"`
	LastHardState     int       `json:"last_hard_state"`
	LastState         int       `json:"last_state"`
	LastStateCritical time.Time `json:"last_state_critical"`
	LastStateOK       time.Time `json:"last_state_ok"`
	LastStateUnknown  time.Time `json:"last_state_unknown"`
	LastStateWarning  time.Time `json:"last_state_warning"`
	State             int       `json:"state"`
}

type Services interface {
	Get(ctx context.Context, name string) (*Service, error)
	Create(ctx context.Context, svc *Service) error
	Delete(ctx context.Context, name string, cascade bool) error
}

// services implements the Services interface.
type services struct {
	ic *client.Icinga
}

// newServicesClient returns a new Services client.
func newServicesClient(cfg *client.Config, log *logr.Logger) *services {
	l := log.WithName("services")
	return &services{ic: client.New(cfg, &l)}
}

// Get returns the service with the given name on the given host.
func (c *services) Get(ctx context.Context, name string) (*Service, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	url := fmt.Sprintf("%s/objects/services/%s", c.ic.Config.BaseURL, name)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.ic.Log.Error(err, "failed creating get-services request")
		return nil, err
	}
	resp, closer, err := c.ic.Call(r) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := closer(); err != nil {
			c.ic.Log.Error(err, "failed closing response body")
		}
	}()

	if resp.StatusCode == http.StatusOK {
		var svc Service
		if err = json.NewDecoder(resp.Body).Decode(&svc); err != nil {
			return nil, err
		}
		return &svc, nil
	}
	c.ic.Log.Error(err, "failed getting service", "name", name)
	return nil, WrapError(resp.Body)
}

// Create creates a new types in Icinga with the given name, if it doesn't already exist.
func (c *services) Create(ctx context.Context, svc *Service) error {
	url := fmt.Sprintf("%s/objects/services/%s", c.ic.Config.BaseURL, svc.Name)

	got, err := c.Get(ctx, svc.Name)
	if got != nil {
		return nil
	}
	if err != nil && !IsNotFound(err) {
		c.ic.Log.Error(err, "failed checking if service exists", "service", svc)
		return err
	}

	p, err := json.Marshal(&CreateObjectRequest[CheckableAttrs]{
		Templates: []string{"generic-service"},
		Attrs:     svc.CheckableAttrs,
	})
	if err != nil {
		c.ic.Log.Error(err, "failed marshalling create-service request")
		return err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(p))
	if err != nil {
		c.ic.Log.Error(err, "failed creating create-service request")
	}
	resp, closer, err := c.ic.Call(r) //nolint:bodyclose
	if err != nil {
		c.ic.Log.Error(err, "failed creating create-service request")
		return err
	}
	defer func() {
		if err := closer(); err != nil {
			c.ic.Log.Error(err, "failed closing response body")
		}
	}()

	// icinga responds with 200 OK if the types was created.
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	c.ic.Log.Error(err, "failed creating the Icinga service", "status", resp.StatusCode, "body", resp.Body)
	return WrapError(resp.Body)
}

// Delete deletes the given service from Icinga.
func (c *services) Delete(ctx context.Context, name string, cascade bool) error {
	url := fmt.Sprintf("%s/objects/services/%s", c.ic.Config.BaseURL, name)
	type deleteServiceRequest struct {
		Cascade bool `json:"cascade"`
	}
	p, err := json.Marshal(&deleteServiceRequest{Cascade: cascade})
	if err != nil {
		c.ic.Log.Error(err, "failed marshalling delete-service request")
		return err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, bytes.NewReader(p))
	if err != nil {
		c.ic.Log.Error(err, "failed creating delete-service request")
		return err
	}
	resp, closer, err := c.ic.Call(r) //nolint:bodyclose
	if err != nil {
		c.ic.Log.Error(err, "failed creating delete-service request")
		return err
	}
	defer func() {
		if err := closer(); err != nil {
			c.ic.Log.Error(err, "failed closing response body")
		}
	}()

	// icinga responds with 200 OK if the service was deleted.
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	c.ic.Log.Error(err, "failed deleting the Icinga service", "status", resp.StatusCode)
	return WrapError(resp.Body)
}
