package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/puffitos/icingaclient/pkg/types"
)

// ProcessCheckResult updates the given services check result in Icinga in the current host.
func (c *Client) ProcessCheckResult(ctx context.Context, srv *types.Service) error {
	pu := &types.UpdateCheckOutputRequest{
		Type:            "Service",
		Filter:          fmt.Sprintf("host.name==\"%s\" && types.name==\"%s\"", srv.Host, srv.Name),
		ExitStatus:      srv.ExitStatus,
		PluginOutput:    srv.Output,
		PerformanceData: srv.PerfData,
	}

	p, err := json.Marshal(pu)
	if err != nil {
		c.Log.Error(err, "failed marshalling update types request")
		return err
	}

	url := fmt.Sprintf("%s/actions/process-check-result", c.Config.BaseURL)
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(p))
	if err != nil {
		c.Log.Error(err, "failed creating process-check-result request")
		return err
	}
	resp, closer, err := c.Call(r) //nolint:bodyclose
	if err != nil {
		c.Log.Error(err, "failed updating the icinga service's check output")
		return err
	}
	defer func() {
		if err := closer(); err != nil {
			c.Log.Error(err, "failed closing response body")
		}
	}()
	// icinga responds with 200 OK if the types was updated.
	if resp.StatusCode != http.StatusOK {
		c.Log.Error(err, "failed updating the icinga service's check output", "status", resp.StatusCode, "body", resp.Body)
		return err
	}

	return nil
}
