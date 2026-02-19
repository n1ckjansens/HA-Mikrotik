package automation

import (
	"strings"
	"time"
)

// CapabilityScope defines where capability state is stored and executed.
type CapabilityScope string

const (
	// ScopeDevice binds capability to a concrete device.
	ScopeDevice CapabilityScope = "device"
	// ScopeGlobal keeps one shared capability state for the whole system.
	ScopeGlobal CapabilityScope = "global"
)

// NormalizeCapabilityScope applies backward-compatible default scope.
func NormalizeCapabilityScope(scope CapabilityScope) CapabilityScope {
	switch CapabilityScope(strings.TrimSpace(string(scope))) {
	case ScopeGlobal:
		return ScopeGlobal
	default:
		return ScopeDevice
	}
}

// ControlType defines frontend UI control type for capability state changes.
type ControlType string

const (
	// ControlSwitch renders a binary switch.
	ControlSwitch ControlType = "switch"
	// ControlSelect renders a multi-state selector.
	ControlSelect ControlType = "select"
)

// ParamFieldKind describes UI field data type.
type ParamFieldKind string

const (
	// ParamString is plain text value.
	ParamString ParamFieldKind = "string"
	// ParamEnum is one-of string values.
	ParamEnum ParamFieldKind = "enum"
	// ParamBool is boolean checkbox value.
	ParamBool ParamFieldKind = "bool"
)

// VisibleIfCondition controls conditional field rendering in UI.
type VisibleIfCondition struct {
	Key    string `json:"key"`
	Equals string `json:"equals"`
}

// ParamField describes one action/state-source parameter for UI forms.
type ParamField struct {
	Key         string              `json:"key"`
	Label       string              `json:"label"`
	Kind        ParamFieldKind      `json:"kind"`
	Required    bool                `json:"required"`
	Description string              `json:"description,omitempty"`
	Options     []string            `json:"options,omitempty"`
	VisibleIf   *VisibleIfCondition `json:"visible_if,omitempty"`
}

// ActionInstance is one runtime action invocation configuration.
type ActionInstance struct {
	ID     string         `json:"id"`
	TypeID string         `json:"type_id"`
	Params map[string]any `json:"params"`
}

// CapabilityStateConfig links a logical state to actions.
type CapabilityStateConfig struct {
	Label          string           `json:"label"`
	ActionsOnEnter []ActionInstance `json:"actions_on_enter"`
}

// CapabilityControlOption is one selectable state option in UI.
type CapabilityControlOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// CapabilityControl describes frontend control model.
type CapabilityControl struct {
	Type    ControlType               `json:"type"`
	Options []CapabilityControlOption `json:"options"`
}

// HAExposeConfig keeps Home Assistant entity export settings.
type HAExposeConfig struct {
	Enabled      bool   `json:"enabled"`
	EntityType   string `json:"entity_type"`
	EntitySuffix string `json:"entity_suffix"`
	NameTemplate string `json:"name_template"`
}

// CapabilitySyncSource describes external truth source.
type CapabilitySyncSource struct {
	TypeID string         `json:"type_id"`
	Params map[string]any `json:"params"`
}

// CapabilitySyncMapping maps external boolean to internal states.
type CapabilitySyncMapping struct {
	WhenTrue  string `json:"when_true"`
	WhenFalse string `json:"when_false"`
}

// CapabilitySyncConfig configures periodic sync behavior.
type CapabilitySyncConfig struct {
	Enabled              bool                  `json:"enabled"`
	Source               CapabilitySyncSource  `json:"source"`
	Mapping              CapabilitySyncMapping `json:"mapping"`
	Mode                 string                `json:"mode"`
	TriggerActionsOnSync bool                  `json:"trigger_actions_on_sync"`
}

// CapabilityTemplate defines reusable automation behavior.
type CapabilityTemplate struct {
	ID           string                           `json:"id"`
	Label        string                           `json:"label"`
	Description  string                           `json:"description"`
	Category     string                           `json:"category"`
	Scope        CapabilityScope                  `json:"scope"`
	Control      CapabilityControl                `json:"control"`
	States       map[string]CapabilityStateConfig `json:"states"`
	DefaultState string                           `json:"default_state"`
	Sync         *CapabilitySyncConfig            `json:"sync,omitempty"`
	HAExpose     HAExposeConfig                   `json:"ha_expose"`
}

// DeviceCapability stores per-device applied state.
type DeviceCapability struct {
	DeviceID     string    `json:"device_id"`
	CapabilityID string    `json:"capability_id"`
	Enabled      bool      `json:"enabled"`
	State        string    `json:"state"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GlobalCapability stores global capability state.
type GlobalCapability struct {
	CapabilityID string `json:"capability_id"`
	Enabled      bool   `json:"enabled"`
	State        string `json:"state"`
}

// CapabilityControlDTO is UI-level control representation.
type CapabilityControlDTO struct {
	Type    ControlType               `json:"type"`
	Options []CapabilityControlOption `json:"options"`
}

// CapabilityUIModel is capability read model returned for one device.
type CapabilityUIModel struct {
	ID          string               `json:"id"`
	Label       string               `json:"label"`
	Description string               `json:"description"`
	Control     CapabilityControlDTO `json:"control"`
	State       string               `json:"state"`
	Enabled     bool                 `json:"enabled"`
}

// CapabilityDeviceAssignment is capability view bound to one device.
type CapabilityDeviceAssignment struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	DeviceIP   string `json:"device_ip,omitempty"`
	Online     bool   `json:"online"`
	Enabled    bool   `json:"enabled"`
	State      string `json:"state"`
}

// ActionExecutionWarning is non-fatal action execution failure detail.
type ActionExecutionWarning struct {
	ActionID string `json:"action_id,omitempty"`
	TypeID   string `json:"type_id"`
	Message  string `json:"message"`
}

// SetStateResult returns state transition outcome.
type SetStateResult struct {
	OK       bool                     `json:"ok"`
	Warnings []ActionExecutionWarning `json:"warnings,omitempty"`
}
