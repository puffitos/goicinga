package api

import (
	"github.com/go-logr/logr"
	"github.com/puffitos/goicinga/pkg/client"
)

// API is the interface that groups all the clients that communicate with the icinga API
type API interface {
	Services() Services
	Actions() Actions
}

// ClientSet is the implementation of the API interface
type ClientSet struct {
	actions  Actions
	hosts    Hosts
	services Services
}

// Services returns the services client
func (c *ClientSet) Services() Services {
	return c.services
}

// Actions returns the actions client
func (c *ClientSet) Actions() Actions {
	return c.actions
}

// NewClientSet creates a new client with the given configuration
func NewClientSet(config *client.Config, log *logr.Logger) *ClientSet {
	return &ClientSet{
		services: newServicesClient(config, log),
		hosts:    newHostsClient(config, log),
		actions:  newActionsClient(config, log),
	}
}
