package types

// UpdateCheckOutputRequest is the request body for updating the check output of a service.
type UpdateCheckOutputRequest struct {
	Type            string   `json:"type"`
	Filter          string   `json:"filter"`
	ExitStatus      int      `json:"exit_status"`
	PluginOutput    string   `json:"plugin_output"`
	PerformanceData []string `json:"performance_data"`
}
