package actions

import (
	"sort"

	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
)

const ActionTypeAddressListMembership = "mikrotik.address_list.set_membership"

type Definition struct {
	Type    domain.ActionType
	Handler Handler
}

var definitions = map[string]Definition{
	ActionTypeAddressListMembership: {
		Type: domain.ActionType{
			ID:          ActionTypeAddressListMembership,
			Label:       "MikroTik: Address-list membership",
			Description: "Add or remove a target value in a MikroTik firewall address-list",
			ParamSchema: []domain.ActionParamField{
				{
					Key:         "list",
					Label:       "Address-list name",
					Kind:        domain.ParamString,
					Required:    true,
					Description: "RouterOS firewall address-list name",
				},
				{
					Key:         "mode",
					Label:       "Mode",
					Kind:        domain.ParamEnum,
					Required:    true,
					Options:     []string{"add", "remove"},
					Description: "Whether to add or remove target from the list",
				},
				{
					Key:         "target",
					Label:       "Target",
					Kind:        domain.ParamEnum,
					Required:    true,
					Options:     []string{"device.ip", "device.mac", "literal_ip"},
					Description: "Source value to apply in address-list",
				},
				{
					Key:       "literal_ip",
					Label:     "Literal IP",
					Kind:      domain.ParamString,
					Required:  true,
					VisibleIf: &domain.VisibleIfCondition{Key: "target", Equals: "literal_ip"},
				},
			},
		},
		Handler: handleAddressListMembership,
	},
}

func ActionTypes() []domain.ActionType {
	out := make([]domain.ActionType, 0, len(definitions))
	for _, definition := range definitions {
		out = append(out, cloneActionType(definition.Type))
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func TypeByID(typeID string) (domain.ActionType, bool) {
	definition, ok := definitions[typeID]
	if !ok {
		return domain.ActionType{}, false
	}
	return cloneActionType(definition.Type), true
}

func HandlerByID(typeID string) (Handler, bool) {
	definition, ok := definitions[typeID]
	if !ok {
		return nil, false
	}
	return definition.Handler, true
}

func cloneActionType(input domain.ActionType) domain.ActionType {
	cloned := input
	cloned.ParamSchema = make([]domain.ActionParamField, len(input.ParamSchema))
	for index := range input.ParamSchema {
		field := input.ParamSchema[index]
		if len(field.Options) > 0 {
			field.Options = append([]string(nil), field.Options...)
		}
		if field.VisibleIf != nil {
			visible := *field.VisibleIf
			field.VisibleIf = &visible
		}
		cloned.ParamSchema[index] = field
	}
	return cloned
}
