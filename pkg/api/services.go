package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/puffitos/goicinga/pkg/client"

	"github.com/go-logr/logr"
)

type Services interface {
	Get(ctx context.Context, host, name string) (*Service, error)
	Create(ctx context.Context, svc *Service) error
	Delete(ctx context.Context, svc *Service) error
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

// Create creates a new types in Icinga with the given name, if it doesn't already exist.
func (c *services) Create(ctx context.Context, svc *Service) error {
	url := fmt.Sprintf("%s/objects/services/%s!%s", c.ic.Config.BaseURL, svc.Host, svc.Name)

	got, err := c.Get(ctx, svc.Host, svc.Name)
	if got != nil {
		return nil
	}
	if err != nil {
		c.ic.Log.Error(err, "failed checking if service exists", "service", svc)
		return err
	}

	p, err := json.Marshal(&CreateServiceRequest{
		Template: []string{"generic-service"},
		Attrs:    svc.Attributes,
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
	if resp.StatusCode != http.StatusOK {
		c.ic.Log.Error(err, "failed creating the Icinga service", "status", resp.StatusCode, "body", resp.Body)
		return err
	}
	return nil
}

// Delete deletes the given service from Icinga.
func (c *services) Delete(ctx context.Context, svc *Service) error {
	url := fmt.Sprintf("%s/objects/services/%s!%s", c.ic.Config.BaseURL, svc.Host, svc.Name)

	r, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		c.ic.Log.Error(err, "failed creating delete-service request")
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
	if resp.StatusCode != http.StatusOK {
		c.ic.Log.Error(err, "failed deleting the Icinga service", "status", resp.StatusCode, "body", resp.Body)
		return err
	}
	return nil
}

// Get returns the service with the given name on the given host.
func (c *services) Get(ctx context.Context, host, name string) (*Service, error) {
	if host == "" || name == "" {
		return nil, fmt.Errorf("service host or name cannot be empty")
	}

	url := fmt.Sprintf("%s/objects/services/%s!%s", c.ic.Config.BaseURL, host, name)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.ic.Log.Error(err, "failed creating get-services request")
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
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	return nil, fmt.Errorf("failed getting service %s on host %s: %s", name, host, resp.Status)
}

const (
	ServiceOk       = 0
	ServiceWarning  = 1
	ServiceCritical = 2
	ServiceUnknown  = 3
)

type Service struct {
	// Name is the name of the service in Icinga.
	Name string
	// Host is the name of the host in Icinga, on which the types is hosted.
	Host string
	// ExitStatus is the exit status of the check.
	ExitStatus int
	// Output is the output of the check.
	Output string
	// PerfData is the performance data of the check.
	PerfData []string
	// Attributes are the attributes of the service.
	Attributes *ServiceAttrs `json:"attrs"`
}

// CreateServiceRequest is the request body for creating a new Services in Icinga.
type CreateServiceRequest struct {
	Template []string      `json:"template"`
	Attrs    *ServiceAttrs `json:"attrs"`
}

// ServiceAttrs are the attributes of a Service instance.
type ServiceAttrs struct {
	Name                string `json:"name"`
	CheckCommand        string `json:"check_command"`
	EnableActiveChecks  bool   `json:"enable_active_checks"`
	EnablePassiveChecks bool   `json:"enable_passive_checks"`
}
