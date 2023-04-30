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

type Actions interface {
	ProcessCheckResult(ctx context.Context, srv *Service) error
}

// actions implements the Actions interface.
type actions struct {
	cs *client.Icinga
}

// newActionsClient returns a new Actions client.
func newActionsClient(cfg *client.Config, log *logr.Logger) Actions {
	l := log.WithName("actions")
	return &actions{cs: client.New(cfg, &l)}
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

	p, err := json.Marshal(pu)
	if err != nil {
		c.cs.Log.Error(err, "failed marshalling update types request")
		return err
	}

	url := fmt.Sprintf("%s/actions/process-check-result", c.cs.Config.BaseURL)
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(p))
	if err != nil {
		c.cs.Log.Error(err, "failed creating process-check-result request")
		return err
	}
	resp, closer, err := c.cs.Call(r) //nolint:bodyclose
	if err != nil {
		c.cs.Log.Error(err, "failed updating the icinga service's check output")
		return err
	}
	defer func() {
		if err := closer(); err != nil {
			c.cs.Log.Error(err, "failed closing response body")
		}
	}()
	// icinga responds with 200 OK if the types was updated.
	if resp.StatusCode != http.StatusOK {
		c.cs.Log.Error(err, "failed updating the icinga service's check output", "status", resp.StatusCode, "body", resp.Body)
		return err
	}

	return nil
}

// UpdateCheckOutputRequest is the request body for updating the check output of a service.
type UpdateCheckOutputRequest struct {
	Type            string   `json:"type"`
	Filter          string   `json:"filter"`
	ExitStatus      int      `json:"exit_status"`
	PluginOutput    string   `json:"plugin_output"`
	PerformanceData []string `json:"performance_data"`
}
