package registry

import (
	"context"
	"testing"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
)

type fakeAction struct {
	id string
}

func (f fakeAction) ID() string { return f.id }

func (f fakeAction) Metadata() automationdomain.ActionMetadata {
	return automationdomain.ActionMetadata{ID: f.id, Label: f.id}
}

func (f fakeAction) Validate(params map[string]any) error { return nil }

func (f fakeAction) Execute(
	ctx context.Context,
	execCtx automationdomain.ActionExecutionContext,
	params map[string]any,
) error {
	return nil
}

type fakeStateSource struct {
	id string
}

func (f fakeStateSource) ID() string { return f.id }

func (f fakeStateSource) Metadata() automationdomain.StateSourceMetadata {
	return automationdomain.StateSourceMetadata{ID: f.id, Label: f.id, OutputType: "boolean"}
}

func (f fakeStateSource) Validate(params map[string]any) error { return nil }

func (f fakeStateSource) Read(
	ctx context.Context,
	sourceCtx automationdomain.StateSourceContext,
	params map[string]any,
) (any, error) {
	return true, nil
}

func TestRegistryRegisterAndLookup(t *testing.T) {
	reg := New()
	reg.RegisterAction(fakeAction{id: "b.action"})
	reg.RegisterAction(fakeAction{id: "a.action"})
	reg.RegisterStateSource(fakeStateSource{id: "z.source"})
	reg.RegisterStateSource(fakeStateSource{id: "x.source"})

	if _, ok := reg.Action("a.action"); !ok {
		t.Fatalf("expected action a.action to be registered")
	}
	if _, ok := reg.StateSource("x.source"); !ok {
		t.Fatalf("expected state source x.source to be registered")
	}

	actionTypes := reg.ActionTypes()
	if len(actionTypes) != 2 {
		t.Fatalf("expected 2 action types, got %d", len(actionTypes))
	}
	if actionTypes[0].ID != "a.action" || actionTypes[1].ID != "b.action" {
		t.Fatalf("unexpected action type order: %+v", actionTypes)
	}

	stateSourceTypes := reg.StateSourceTypes()
	if len(stateSourceTypes) != 2 {
		t.Fatalf("expected 2 state source types, got %d", len(stateSourceTypes))
	}
	if stateSourceTypes[0].ID != "x.source" || stateSourceTypes[1].ID != "z.source" {
		t.Fatalf("unexpected state source order: %+v", stateSourceTypes)
	}
}
