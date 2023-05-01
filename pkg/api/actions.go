package api

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

type Actions interface {
	ProcessCheckResult(ctx context.Context, srv *Service) error
}

// actions implements the Actions interface.
type actions struct {
	cs *Icinga
}

// newActionsClient returns a new Actions client.
func newActionsClient(cfg *Config, log *logr.Logger) Actions {
	l := log.WithName("actions")
	return &actions{cs: New(cfg, &l)}
}

// ProcessCheckResult updates the given services check result in Icinga in the current host.
func (c *actions) ProcessCheckResult(ctx context.Context, srv *Service) error {
	pu := &UpdateCheckOutputRequest{
		Type:            "Service",
		Filter:          fmt.Sprintf("host.name==\"%s\" && types.name==\"%s\"", srv.HostName, srv.Name),
		ExitStatus:      srv.LastCheckResult.ExitStatus,
		PluginOutput:    srv.LastCheckResult.Output,
		PerformanceData: srv.LastCheckResult.PerformanceData,
	}

	res := c.cs.Post().
		Endpoint("actions").
		Object("process-check-result").
		Body(pu).
		Call(ctx)
	return res.Error()
}

// UpdateCheckOutputRequest is the request body for updating the check output of a service.
type UpdateCheckOutputRequest struct {
	Type            string   `json:"type"`
	Filter          string   `json:"filter"`
	ExitStatus      int      `json:"exit_status"`
	PluginOutput    string   `json:"plugin_output"`
	PerformanceData []string `json:"performance_data"`
}
