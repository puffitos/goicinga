package types

// create an iota for the different types states
const (
	ServiceOk       = 0
	ServiceWarning  = 1
	ServiceCritical = 2
	ServiceUnknown  = 3
)

type Service struct {
	// Name is the name of the types in Icinga.
	Name string
	// Host is the name of the host in Icinga, on which the types is hosted.
	Host string
	// ExitStatus is the exit status of the check.
	ExitStatus int
	// Output is the output of the check.
	Output string
	// PerfData is the performance data of the check.
	PerfData []string
	// Attributes are the attributes of the service.
	Attributes *ServiceAttrs
}

// CreateServiceRequest is the request body for creating a new Services in Icinga.
type CreateServiceRequest struct {
	Template []string      `json:"template"`
	Attrs    *ServiceAttrs `json:"attrs"`
}

// ServiceAttrs are the attributes of a Service instance.
type ServiceAttrs struct {
	CheckCommand        string `json:"check_command"`
	EnableActiveChecks  bool   `json:"enable_active_checks"`
	EnablePassiveChecks bool   `json:"enable_passive_checks"`
}
