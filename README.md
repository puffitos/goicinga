# Icinga2 API Go Client

A Go client for the Icinga2 API, heavily inspired from the kubernetes client-go library. Marrying the old world with the
new one, one library at a time.

## Usage

Create a `ClientSet` to be more flexible (allows access to multiple Icinga2 resources):

```go
import "github.com/puffitos/goicinga/pkg/api"

// initialize a ClientSet with an api.Config and a logger (optional)
// and get the host, ignoring the error for brevity
cs := api.NewClientSet(cfg, log)
host, _ := cs.Hosts().Get("my-host")
fmt.Sprintf("Host: %s", host.Name)
```

Or use the `Icinga` client directly, to only do what you need:

```go
import "github.com/puffitos/goicinga/pkg/api"

// initialize the client with an api.Config and a logger (optional)
client := api.New(cfg, log)
ctx := context.Background()

// Get the host, ignoring the error for brevity
host, _ := client.Get().
        Endpoint("objects").
        Type("hosts").
        Object("my-host").
        Call(ctx).
        Into(&res)

fmt.Sprintf("Host: %s", host.Name)
```

## Development

Run `make setup-icinga` to run a local Icinga2 instance in a Docker container. The password of the root user can be
found in the icinga-master container, under the path `etc/icinga2/conf.d/api-users.conf`.