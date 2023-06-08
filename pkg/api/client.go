package api

import (
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/puffitos/goicinga/internal/util"
)

type Client interface {
	Get() *Request
	Post() *Request
	Put() *Request
	Delete() *Request
	Verb(verb string) *Request
}

// Icinga handles the HTTP communication with the icinga API
type Icinga struct {
	Config *Config
	Client *http.Client
	Log    *logr.Logger
}

// Config holds the configuration for the icinga client
type Config struct {
	// BaseURL the URL of the icinga API (including the port)
	BaseURL string
	// APIUser is the username for the icinga API
	APIUser string
	// APIPass is the password for the icinga API
	APIPass string
	// Timeout is the global timeout for all API requests
	Timeout time.Duration
	// CertPath is the endpoint to the certificate used for TLS
	CertPath string
}

// New creates a new icinga client with the passed configuration and logger
func New(config *Config, log *logr.Logger) *Icinga {
	if log == nil {
		l := logr.Discard()
		log = &l
	}

	return &Icinga{
		Config: config,
		Client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: util.NewTLSConfig(config.CertPath),
			},
		},
		Log: log,
	}
}

// Verb creates a new request for the given verb.
func (c *Icinga) Verb(verb string) *Request {
	r := NewRequest(c)
	r.verb = verb
	return r
}

// Get creates a new GET request.
func (c *Icinga) Get() *Request {
	return c.Verb(http.MethodGet)
}

// Post creates a new POST request.
func (c *Icinga) Post() *Request {
	return c.Verb(http.MethodPost)
}

// Put creates a new PUT request.
func (c *Icinga) Put() *Request {
	return c.Verb(http.MethodPut)
}

// Delete creates a new DELETE request.
func (c *Icinga) Delete() *Request {
	return c.Verb(http.MethodDelete)
}
