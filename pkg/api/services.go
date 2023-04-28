package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

	p, err := json.Marshal(&CreateObjectRequest[ServiceAttrs]{
		Templates: []string{"generic-service"},
		Attrs:     svc.Attributes,
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

// ServiceAttrs are the attributes of a Service instance.
type ServiceAttrs struct {
	// The service name. Must be unique on a per-host basis. For advanced usage in apply rules only.
	Name string `json:"name"`
	// A short description of the service.
	DisplayName string `json:"display_name,omitempty"`
	// The host this service belongs to. There must be a Host object with that name.
	HostName string `json:"host_name"`
	// The service groups this service belongs to.
	Groups []string `json:"groups,omitempty"`
	// A map containing custom variables that are specific to this service.
	Vars map[string]interface{} `json:"vars,omitempty"`
	// The name of the check command.
	CheckCommand string `json:"check_command"`
	// The number of times a service is re-checked before changing into a hard state. Defaults to 3.
	MaxCheckAttempts int `json:"max_check_attempts,omitempty"`
	// The name of a time period which determines when this service should be checked.
	// Not set by default (effectively 24x7).
	CheckPeriod string `json:"check_period,omitempty"`
	// Check command timeout in seconds. Overrides the CheckCommand’s timeout attribute
	CheckTimeout time.Duration `json:"check_timeout,omitempty"`
	// The check interval (in seconds). This interval is used for
	// checks when the service is in a HARD state. Defaults to 5m.
	CheckInterval time.Duration `json:"check_interval,omitempty"`
	// The retry interval (in seconds). This interval is used for checks
	// when the service is in a SOFT state. Defaults to 1m.
	// Note: This does not affect the scheduling after a passive check result.
	RetryInterval time.Duration `json:"retry_interval,omitempty"`
	// Whether notifications are enabled. Defaults to true.
	EnableNotifications bool `json:"enable_notifications,omitempty"`
	// Whether active checks are enabled. Defaults to true.
	EnableActiveChecks bool `json:"enable_active_checks,omitempty"`
	// Whether passive checks are enabled. Defaults to true.
	EnablePassiveChecks bool `json:"enable_passive_checks,omitempty"`
	// Enables event handlers for this host. Defaults to true.
	EnableEventHandler bool `json:"enable_event_handler,omitempty"`
	// Whether flap detection is enabled. Defaults to false.
	EnableFlapping bool `json:"enable_flapping,omitempty"`
	// Flapping upper bound in percent for a service to be considered flapping. 30.0
	FlappingThresholdHigh float64 `json:"flapping_threshold_high,omitempty"`
	// Flapping lower bound in percent for a service to be considered not flapping. 25.0
	FlappingThresholdLow float64 `json:"flapping_threshold_low,omitempty"`
	// A list of states that should be ignored during flapping calculation. By default, no state is ignored.
	FlappingIgnoreStates []int `json:"flapping_ignore_states,omitempty"`
	// Whether performance data processing is enabled. Defaults to true.
	EnablePerfData bool `json:"enable_perfdata,omitempty"`
	// The name of an event command that should be executed every time
	// the service’s state changes or the service is in a SOFT state.
	EventCommand string `json:"event_command,omitempty"`
	// Treat all state changes as HARD changes. See here for details. Defaults to false.
	Volatile bool `json:"volatile,omitempty"`
	// The zone this object is a member of. Please read the distributed monitoring chapter for details.
	Zone string `json:"zone,omitempty"`
	// The endpoint where commands are executed on.
	CommandEndpoint string `json:"command_endpoint,omitempty"`
	// Notes for the service.
	Notes string `json:"notes,omitempty"`
	// URL for notes for the service (for example, in notification commands).
	NotesURL string `json:"notes_url,omitempty"`
	// URL for actions for the service (for example, an external graphing tool).
	ActionURL string `json:"action_url,omitempty"`
	// Icon image for the service. Used by external interfaces only.
	IconImage string `json:"icon_image,omitempty"`
	// Icon image description for the service. Used by external interface only.
	IconImageAlt string `json:"icon_image_alt,omitempty"`
}
