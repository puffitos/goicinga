package api

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
