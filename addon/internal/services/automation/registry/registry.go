package registry

import (
	"sort"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
)

// Registry stores action and state-source implementations.
type Registry struct {
	actions      map[string]automationdomain.Action
	stateSources map[string]automationdomain.StateSource
}

// New creates empty action/state-source registry.
func New() *Registry {
	return &Registry{
		actions:      map[string]automationdomain.Action{},
		stateSources: map[string]automationdomain.StateSource{},
	}
}

// RegisterAction adds action implementation by ID.
func (r *Registry) RegisterAction(a automationdomain.Action) {
	if a == nil {
		return
	}
	r.actions[a.ID()] = a
}

// RegisterStateSource adds state-source implementation by ID.
func (r *Registry) RegisterStateSource(s automationdomain.StateSource) {
	if s == nil {
		return
	}
	r.stateSources[s.ID()] = s
}

// Action returns registered action by ID.
func (r *Registry) Action(id string) (automationdomain.Action, bool) {
	a, ok := r.actions[id]
	return a, ok
}

// StateSource returns registered state-source by ID.
func (r *Registry) StateSource(id string) (automationdomain.StateSource, bool) {
	s, ok := r.stateSources[id]
	return s, ok
}

// ActionTypes returns stable action metadata list for HTTP DTO.
func (r *Registry) ActionTypes() []automationdomain.ActionMetadata {
	out := make([]automationdomain.ActionMetadata, 0, len(r.actions))
	for _, item := range r.actions {
		out = append(out, item.Metadata())
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

// StateSourceTypes returns stable state-source metadata list for HTTP DTO.
func (r *Registry) StateSourceTypes() []automationdomain.StateSourceMetadata {
	out := make([]automationdomain.StateSourceMetadata, 0, len(r.stateSources))
	for _, item := range r.stateSources {
		out = append(out, item.Metadata())
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}
