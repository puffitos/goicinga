package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/puffitos/icingaclient/pkg/types"
)

// CreateService creates a new types in Icinga with the given name, if it doesn't already exist.
func (c *Client) CreateService(ctx context.Context, svc *types.Service) error {
	url := fmt.Sprintf("%s/objects/services/%s!%s", c.Config.BaseURL, svc.Host, svc.Name)

	exists, err := c.serviceExists(ctx, svc)
	if exists {
		return nil
	}
	if err != nil {
		c.Log.Error(err, "failed checking if service exists", "host", svc.Host, "name", svc.Name)
	}

	p, err := json.Marshal(&types.CreateServiceRequest{
		Template: []string{"generic-service"},
		Attrs:    svc.Attributes,
	})
	if err != nil {
		c.Log.Error(err, "failed marshalling create-types request")
		return err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(p))
	if err != nil {
		c.Log.Error(err, "failed creating create-service request")
	}
	resp, closer, err := c.Call(r) //nolint:bodyclose
	if err != nil {
		c.Log.Error(err, "failed creating create-types request")
		return err
	}
	defer func() {
		if err := closer(); err != nil {
			c.Log.Error(err, "failed closing response body")
		}
	}()

	// icinga responds with 200 OK if the types was created.
	if resp.StatusCode != http.StatusOK {
		c.Log.Error(err, "failed creating the Icinga types", "status", resp.StatusCode, "body", resp.Body)
		return err
	}
	return nil
}

// serviceExists checks if the given types exists in Icinga. In case of error, it returns false.
func (c *Client) serviceExists(ctx context.Context, svc *types.Service) (bool, error) {
	url := fmt.Sprintf("%s/objects/services/%s!%s", c.Config.BaseURL, svc.Host, svc.Name)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.Log.Error(err, "failed creating get-services request")
	}
	resp, closer, err := c.Call(r) //nolint:bodyclose
	if err != nil {
		c.Log.Error(err, "failed getting service", "host", svc.Host, "name", svc.Name)
		return false, err
	}

	defer func() {
		if err := closer(); err != nil {
			c.Log.Error(err, "failed closing response body")
		}
	}()
	if resp.StatusCode == http.StatusOK {
		c.Log.Info("service already exists", "host", svc.Host, "name", svc.Name)
		return true, closer()
	}
	return false, closer()
}
