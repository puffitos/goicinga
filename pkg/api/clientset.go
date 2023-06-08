package api

import (
	"github.com/go-logr/logr"
)

// API is the interface that groups all the clients that communicate with the icinga API
type API interface {
	Services() Services
	Hosts() Hosts
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

// Hosts returns the hosts client
func (c *ClientSet) Hosts() Hosts {
	return c.hosts
}

// Actions returns the actions client
func (c *ClientSet) Actions() Actions {
	return c.actions
}

// NewClientSet creates a new client with the given configuration
func NewClientSet(config *Config, log *logr.Logger) *ClientSet {
	if log == nil {
		l := logr.Discard()
		log = &l
	}

	return &ClientSet{
		services: newServicesClient(config, log),
		hosts:    newHostsClient(config, log),
		actions:  newActionsClient(config, log),
	}
}
