package api

import (
	"encoding/json"
	"time"
)

const Ms = 1e9

// Object Hierarchy in Icinga2, mimicked by embedding structs
// Object -> ConfigObject -> CustomVar -> Checkable -> Host/Service

// Attributes represents the attributes of a creatable icinga object.
type Attributes interface {
	CheckableAttrs | ConfigObjectAttrs | ObjectAttrs | CustomVarAttrs
}

// ObjectAttrs represents the attributes of an icinga object.
type ObjectAttrs struct {
	Type string `json:"type"`
}

// ConfigObjectAttrs contains the attributes of a config object.
type ConfigObjectAttrs struct {
	ObjectAttrs
	Name               string                 `json:"name"`
	Active             bool                   `json:"active"`
	Extensions         map[string]interface{} `json:"extensions"`
	HAMode             int                    `json:"ha_mode"`
	OriginalAttributes map[string]interface{} `json:"original_attributes"`
	Package            string                 `json:"package"`
	PauseCalled        bool                   `json:"pause_called"`
	Paused             bool                   `json:"paused"`
	ResumeCalled       bool                   `json:"resume_called"`
	SourceLocation     map[string]interface{} `json:"source_location"`
	StartCalled        bool                   `json:"start_called"`
	StateLoaded        bool                   `json:"state_loaded"`
	StopCalled         bool                   `json:"stop_called"`
	Templates          []string               `json:"templates"`
	Version            string                 `json:"version"`
	Zone               string                 `json:"zone"`
}

// CustomVarAttrs represents the custom variable attributes of an icinga object.
type CustomVarAttrs struct {
	ConfigObjectAttrs
	Vars map[string]interface{} `json:"vars,omitempty"`
}

// CheckableAttrs represents the checkable attributes of an icinga object.
type CheckableAttrs struct {
	CustomVarAttrs
	// The name of the check command.
	CheckCommand string `json:"check_command"`
	// The number of times an object is re-checked before changing into a hard state. Defaults to 3.
	MaxCheckAttempts int `json:"max_check_attempts,omitempty"`
	// The name of a time period which determines when this object should be checked.
	// Not set by default (effectively 24x7).
	CheckPeriod string `json:"check_period,omitempty"`
	// Check command timeout in seconds. Overrides the CheckCommand’s timeout attribute
	CheckTimeout time.Duration `json:"check_timeout,omitempty"`
	// The check interval (in seconds). This interval is used for
	// checks when the object is in a HARD state. Defaults to 5m.
	CheckInterval time.Duration `json:"check_interval,omitempty"`
	// The retry interval (in seconds). This interval is used for checks
	// when the object is in a SOFT state. Defaults to 1m.
	// Note: This does not affect the scheduling after a passive check result.
	RetryInterval time.Duration `json:"retry_interval,omitempty"`
	// Whether notifications are enabled. Defaults to true.
	EnableNotifications bool `json:"enable_notifications,omitempty"`
	// Whether active checks are enabled. Defaults to true.
	EnableActiveChecks bool `json:"enable_active_checks,omitempty"`
	// Whether passive checks are enabled. Defaults to true.
	EnablePassiveChecks bool `json:"enable_passive_checks,omitempty"`
	// Enables event handlers for this host. Defaults to true.
	EnableEventHandler bool `json:"enable_event_handler,omitempty"`
	// Whether flap detection is enabled. Defaults to false.
	EnableFlapping bool `json:"enable_flapping,omitempty"`
	// Flapping upper bound in percent for a object to be considered flapping. 30.0
	FlappingThresholdHigh float64 `json:"flapping_threshold_high,omitempty"`
	// Flapping lower bound in percent for a object to be considered not flapping. 25.0
	FlappingThresholdLow float64 `json:"flapping_threshold_low,omitempty"`
	// A list of states that should be ignored during flapping calculation. By default, no state is ignored.
	FlappingIgnoreStates []int `json:"flapping_ignore_states,omitempty"`
	// Whether performance data processing is enabled. Defaults to true.
	EnablePerfData bool `json:"enable_perfdata,omitempty"`
	// The name of an event command that should be executed every time
	// the object’s state changes or the object is in a SOFT state.
	EventCommand string `json:"event_command,omitempty"`
	// Treat all state changes as HARD changes. See here for details. Defaults to false.
	Volatile bool `json:"volatile,omitempty"`
	// The zone this object is a member of. Please read the distributed monitoring chapter for details.
	Zone string `json:"zone,omitempty"`
	// The endpoint where commands are executed on.
	CommandEndpoint string `json:"command_endpoint,omitempty"`
	// Notes for the object.
	Notes string `json:"notes,omitempty"`
	// URL for notes for the object (for example, in notification commands).
	NotesURL string `json:"notes_url,omitempty"`
	// URL for actions for the object (for example, an external graphing tool).
	ActionURL string `json:"action_url,omitempty"`
	// Icon image for the object. Used by external interfaces only.
	IconImage string `json:"icon_image,omitempty"`
	// Icon image description for the object. Used by external interface only.
	IconImageAlt string `json:"icon_image_alt,omitempty"`
	// The Acknowledgement type.
	Acknowledgement Acknowledgement `json:"acknowledgement,omitempty"`
	// When the acknowledgement expires (as a UNIX timestamp; 0 = no expiry).
	AcknowledgementExpiry time.Time `json:"acknowledgement_expiry,omitempty"`
	// When the acknowledgement has been set/cleared
	AcknowledgementLastChange time.Time `json:"acknowledgement_last_change,omitempty"`
	// The current check attempt number.
	CheckAttempt int `json:"check_attempt,omitempty"`
	// Whether the service has one or more active downtimes.
	DowntimeDepth int `json:"downtime_depth,omitempty"`
	// The execution times for the check. Read-only.
	Executions map[string]interface{} `json:"executions,omitempty"`
	// Whether the object is flapping between states
	Flapping bool `json:"flapping,omitempty"`
	// Current flapping value in percent (see flapping_thresholds)
	FlappingCurrent float64 `json:"flapping_current,omitempty"`
	// When the last flapping change occurred.
	FlappingLastChange time.Time `json:"flapping_last_change,omitempty"`
	// Whether next check is forced.
	ForceNextCheck bool `json:"force_next_check,omitempty"`
	// Whether next notification is forced.
	ForceNextNotification bool `json:"force_next_notification,omitempty"`
	// Whether the problem is handled (downtime or acknowledgement).
	Handled bool `json:"handled,omitempty"`
	// When the last check occurred.
	LastCheck time.Time `json:"last_check,omitempty"`
	// The current CheckResult.
	LastCheckResult CheckResult `json:"last_check_result,omitempty"`
	// When the last hard state change occurred.
	LastHardStateChange time.Time `json:"last_hard_state_change,omitempty"`
	// Whether the service was reachable when the last check occurred.
	LastReachable bool `json:"last_reachable,omitempty"`
	// When the last state change occurred.
	LastStateChange time.Time `json:"last_state_change,omitempty"`
	// The previous StateType.
	LastStateType StateType `json:"last_state_type,omitempty"`
	// When the object was unreachable the last time.
	LastStateUnreachable time.Time `json:"last_state_unreachable,omitempty"`
	// When the next check occurs.
	NextCheck time.Time `json:"next_check,omitempty"`
	// When the next check update is to be expected.
	NextUpdate time.Time `json:"next_update,omitempty"`
	// Previous timestamp of last_state_change before processing a new check result.
	PreviousStateChange time.Time `json:"previous_state_change,omitempty"`
	// Whether the object is considered in a problem state type (NOT-OK / NOT-UP).
	Problem bool `json:"problem,omitempty"`
	// Calculated value of severity (https://icinga.com/docs/icinga-2/latest/doc/19-technical-concepts/#severity).
	Severity int `json:"severity,omitempty"`
	// The current StateType.
	StateType StateType `json:"state_type,omitempty"`
}

// CheckResult represents the results of a Service check.
type CheckResult struct {
	// Scheduled check execution start time.
	ScheduleStart time.Time `json:"schedule_start,omitempty"`
	// Scheduled check execution end time.
	ScheduleEnd time.Time `json:"schedule_end,omitempty"`
	// Actual check execution start time.
	ExecutionStart time.Time `json:"execution_start,omitempty"`
	// Actual check execution end time.
	ExecutionEnd time.Time `json:"execution_end,omitempty"`
	// Array of command with shell-escaped arguments or command line string.
	Command []string `json:"command,omitempty"`
	// The exit status returned by the check execution.
	ExitStatus int `json:"exit_status,omitempty"`
	// The current ServiceState / HostState.
	State int `json:"state,omitempty"`
	// The previous hard ServiceState / HostState.
	PreviousHardState int `json:"previous_hard_state,omitempty"`
	// The check output.
	Output string `json:"output,omitempty"`
	// Array of performance data values.
	PerformanceData []string `json:"performance_data,omitempty"`
	// Whether the result is from an active or passive check.
	Active bool `json:"active,omitempty"`
	// Name of the node executing the check.
	CheckSource string `json:"check_source,omitempty"`
	// Name of the node scheduling the check.
	SchedulingSource string `json:"scheduling_source,omitempty"`
	// Time-to-live duration in seconds for this check result.
	// The next expected check result is now + ttl where freshness checks are executed.
	TTL float64 `json:"ttl,omitempty"`
	// Internal attribute used for calculations.
	VarsBefore map[string]interface{} `json:"vars_before,omitempty"`
	// Internal attribute used for calculations.
	VarsAfter map[string]interface{} `json:"vars_after,omitempty"`
}

type Acknowledgement int

const (
	None Acknowledgement = iota
	Normal
	Sticky
)

type StateType int

const (
	StateTypeSoft StateType = iota
	StateTypeHard
)

// ObjectQuery represents a query for an Icinga object.
type ObjectQuery struct {
	Attrs      []string               `json:"attrs,omitempty"`
	Filter     string                 `json:"filter,omitempty"`
	FilterVars map[string]interface{} `json:"filter_vars,omitempty"`
	Joins      []string               `json:"joins,omitempty"`
}

// CreateObjectRequest is the request body for creating a new config object in icinga.
// T represents the type of the attributes of the config object to create.
type CreateObjectRequest[T Attributes] struct {
	Templates      []string `json:"templates,omitempty"`
	Attrs          T        `json:"attrs"`
	IgnoredOnError bool     `json:"ignore_on_error,omitempty"`
}

// deleteObjectRequest is the request body for deleting a config object in icinga.
type deleteObjectRequest struct {
	Cascade bool `json:"cascade"`
}

// ObjectQueryResult represents the api representation of a single icinga object.
type ObjectQueryResult struct {
	Name  string                 `json:"name"`
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs"`
	Joins map[string]interface{} `json:"joins"`
	Meta  map[string]interface{} `json:"meta"`
}

// ObjectQueryResults represents the results of a query to the icinga Objects endpoint.
type ObjectQueryResults struct {
	Results []ObjectQueryResult `json:"results"`
}

func (r *ObjectQueryResult) MarshalJSON() ([]byte, error) {
	type Alias ObjectQueryResult
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *ObjectQueryResult) UnmarshalJSON(data []byte) error {
	type Alias ObjectQueryResult
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}
