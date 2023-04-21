package client

import (
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/puffitos/icingaclient/internal/util"
)

// Client handles the communication with the icinga API
type Client struct {
	Config *Config
	Conn   *http.Client
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
	// CertPath is the path to the certificate used for TLS
	CertPath string
}

// NewClient creates a new client with the given configuration
func NewClient(config *Config, log *logr.Logger) *Client {
	tlsCfg := util.NewTLSConfig(config.CertPath)
	return &Client{
		Config: config,
		Conn: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: tlsCfg,
			},
		},
		Log: log,
	}
}

// Call executes the given request and returns an error if the request failed.
// The calling function is expected to close the response body using the returned function.
func (c *Client) Call(req *http.Request) (*http.Response, func() error, error) {
	c.Log.V(1).Info("calling icinga api", "url", req.URL.String(), "method", req.Method, "body", req.Body)
	req.SetBasicAuth(c.Config.APIUser, c.Config.APIPass)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-HTTP-Method-Override", req.Method)

	resp, err := c.Conn.Do(req)
	if err != nil {
		c.Log.Error(err, "failed to call icinga api", "path", req.URL.Path)
		return nil, nil, err
	}

	// Return the response along with a function that can be used to close the response body
	closer := func() error {
		if resp.Body != nil {
			return resp.Body.Close()
		}
		return nil
	}
	c.Log.V(1).Info("response from icinga api", "status", resp.StatusCode)
	return resp, closer, nil
}
