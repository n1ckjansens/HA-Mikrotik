package domain

import "time"

type ActionParamFieldKind string

const (
	ParamString ActionParamFieldKind = "string"
	ParamEnum   ActionParamFieldKind = "enum"
	ParamBool   ActionParamFieldKind = "bool"
)

type VisibleIfCondition struct {
	Key    string `json:"key"`
	Equals string `json:"equals"`
}

type ActionParamField struct {
	Key         string               `json:"key"`
	Label       string               `json:"label"`
	Kind        ActionParamFieldKind `json:"kind"`
	Required    bool                 `json:"required"`
	Description string               `json:"description,omitempty"`
	Options     []string             `json:"options,omitempty"`
	VisibleIf   *VisibleIfCondition  `json:"visible_if,omitempty"`
}

type ActionType struct {
	ID          string             `json:"id"`
	Label       string             `json:"label"`
	Description string             `json:"description"`
	ParamSchema []ActionParamField `json:"param_schema"`
}

type ActionInstance struct {
	ID     string         `json:"id"`
	TypeID string         `json:"type_id"`
	Params map[string]any `json:"params"`
}

type ControlType string

const (
	ControlSwitch ControlType = "switch"
	ControlSelect ControlType = "select"
)

type CapabilityControlOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type CapabilityControl struct {
	Type    ControlType               `json:"type"`
	Options []CapabilityControlOption `json:"options"`
}

type CapabilityStateConfig struct {
	Label          string           `json:"label"`
	ActionsOnEnter []ActionInstance `json:"actions_on_enter"`
}

type CapabilitySyncSource struct {
	TypeID string         `json:"type_id"`
	Params map[string]any `json:"params"`
}

type CapabilitySyncMapping struct {
	WhenTrue  string `json:"when_true"`
	WhenFalse string `json:"when_false"`
}

type CapabilitySyncConfig struct {
	Enabled              bool                  `json:"enabled"`
	Source               CapabilitySyncSource  `json:"source"`
	Mapping              CapabilitySyncMapping `json:"mapping"`
	Mode                 string                `json:"mode"`
	TriggerActionsOnSync bool                  `json:"trigger_actions_on_sync"`
}

type HAExposeConfig struct {
	Enabled      bool   `json:"enabled"`
	EntityType   string `json:"entity_type"`
	EntitySuffix string `json:"entity_suffix"`
	NameTemplate string `json:"name_template"`
}

type CapabilityTemplate struct {
	ID           string                           `json:"id"`
	Label        string                           `json:"label"`
	Description  string                           `json:"description"`
	Category     string                           `json:"category"`
	Control      CapabilityControl                `json:"control"`
	States       map[string]CapabilityStateConfig `json:"states"`
	DefaultState string                           `json:"default_state"`
	Sync         *CapabilitySyncConfig            `json:"sync,omitempty"`
	HAExpose     HAExposeConfig                   `json:"ha_expose"`
}

type DeviceCapabilityState struct {
	DeviceID     string    `json:"device_id"`
	CapabilityID string    `json:"capability_id"`
	Enabled      bool      `json:"enabled"`
	State        string    `json:"state"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CapabilityControlDTO struct {
	Type    ControlType               `json:"type"`
	Options []CapabilityControlOption `json:"options"`
}

type CapabilityUIModel struct {
	ID          string               `json:"id"`
	Label       string               `json:"label"`
	Description string               `json:"description"`
	Control     CapabilityControlDTO `json:"control"`
	State       string               `json:"state"`
	Enabled     bool                 `json:"enabled"`
}

type CapabilityDeviceAssignment struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	DeviceIP   string `json:"device_ip,omitempty"`
	Online     bool   `json:"online"`
	Enabled    bool   `json:"enabled"`
	State      string `json:"state"`
}

type ActionExecutionWarning struct {
	ActionID string `json:"action_id,omitempty"`
	TypeID   string `json:"type_id"`
	Message  string `json:"message"`
}

type SetStateResult struct {
	OK       bool                     `json:"ok"`
	Warnings []ActionExecutionWarning `json:"warnings,omitempty"`
}
