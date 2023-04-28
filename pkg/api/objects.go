package api

// ObjectQuery represents a query to the icinga Objects endpoint.
type ObjectQuery struct {
	Attrs      []string               `json:"attrs,omitempty"`
	Filter     string                 `json:"filter,omitempty"`
	FilterVars map[string]interface{} `json:"filter_vars,omitempty"`
	Joins      []string               `json:"joins,omitempty"`
}

// ObjectQueryResult represents the result of a query to the icinga Objects endpoint.
type ObjectQueryResult struct {
	Name  string                 `json:"name"`
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs"`
	Joins map[string]interface{} `json:"joins"`
	Meta  map[string]interface{} `json:"meta"`
}

// ObjectAttrs represents the attributes of an icinga object.
type ObjectAttrs interface {
	ServiceAttrs
}

// CreateObjectRequest is the request body for creating a new config object in Icinga.
// T represents the type of the attributes of the config object to create.
type CreateObjectRequest[T ObjectAttrs] struct {
	Templates      []string `json:"templates,omitempty"`
	Attrs          *T       `json:"attrs"`
	IgnoredOnError bool     `json:"ignore_on_error,omitempty"`
}
