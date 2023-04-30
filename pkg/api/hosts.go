package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/puffitos/goicinga/pkg/client"
)

type Host struct {
	// A short description of the host (e.g. displayed by external interfaces instead of the name if set).
	DisplayName string `json:"display_name,omitempty"`
	// The host’s IPv4 address. Available as command runtime macro $address$ if set.
	Address string `json:"address,omitempty"`
	// The host’s IPv6 address. Available as command runtime macro $address6$ if set.
	Address6 string `json:"address_6,omitempty"`
	// A list of host groups this host belongs to.
	Groups []string `json:"groups,omitempty"`
	CheckableAttrs
}

type HostState int

const (
	HostUp HostState = iota
	HostDown
)

// Hosts is the interface for interacting with Icinga hosts.
type Hosts interface {
	Get(ctx context.Context, name string) (*Host, error)
}

// hosts implements the Hosts interface.
type hosts struct {
	ic *client.Icinga
}

// newHostsClient returns a new Hosts client.
func newHostsClient(cfg *client.Config, log *logr.Logger) *hosts {
	l := log.WithName("hosts")
	return &hosts{ic: client.New(cfg, &l)}
}

// Get returns the host with the given name, or nil if it doesn't exist.
func (c *hosts) Get(ctx context.Context, name string) (*Host, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	url := fmt.Sprintf("%s/objects/hosts/%s", c.ic.Config.BaseURL, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.ic.Log.Error(err, "failed creating get-host request")
		return nil, err
	}

	resp, closer, err := c.ic.Call(req) //nolint:bodyclose
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := closer(); err != nil {
			c.ic.Log.Error(err, "failed closing response body")
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return nil, nil
}
