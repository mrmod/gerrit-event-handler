package main

import (
	"context"
	"testing"

	"github.com/mrmod/gerrit-buildkite/backend"
)

func TestItSavesBuildWhenAPatchsetIsCreated(t *testing.T) {
	p := NewMockPipeline()
	b := NewMockBackend()

	event := Event{
		Type: "patchset-created",
		PatchSet: PatchSet{
			Number:   1,
			Revision: "123456",
		},
		Change: Change{
			Number: 9999,
		},
	}
	if err := HandlePatchsetCreated(event, p, b); err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if p.FunctionCallCounter["CreateBuild"] != 1 {
		t.Errorf("Expected CreateBuild to be called once, but it was called %d times", p.FunctionCallCounter["CreateBuild"])
	}

	if b.FunctionCallCounter["SaveBuild"] != 1 {
		t.Errorf("Expected SaveBuild to be called once, but it was called %d times", p.FunctionCallCounter["SaveBuild"])
	}
	if p.FunctionCallCounter["CancelBuild"] != 0 {
		t.Errorf("Expected CancelBuild to be called zero times, but it was called %d times", p.FunctionCallCounter["CancelBuild"])
	}
}

func TestItCancelsThePreviousPatchSet(t *testing.T) {
	p := NewMockPipeline()
	b := NewMockBackend()
	b.MockGetPatch = func(ctx context.Context, patch *backend.Patch) (*backend.PatchBuild, error) {
		return &backend.PatchBuild{
			BuildNumber: 123,
			Patch:       patch,
		}, nil
	}
	event := Event{
		Type: "patchset-created",
		PatchSet: PatchSet{
			Number:   2,
			Revision: "123456",
		},
		Change: Change{
			Number: 9999,
		},
	}
	if err := HandlePatchsetCreated(event, p, b); err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if p.FunctionCallCounter["CreateBuild"] != 1 {
		t.Errorf("Expected CreateBuild to be called once, but it was called %d times", p.FunctionCallCounter["CreateBuild"])
	}

	if b.FunctionCallCounter["SaveBuild"] != 1 {
		t.Errorf("Expected SaveBuild to be called once, but it was called %d times", p.FunctionCallCounter["SaveBuild"])
	}
	if p.FunctionCallCounter["CancelBuild"] != 1 {
		t.Errorf("Expected CancelBuild to be called once, but it was called %d times", p.FunctionCallCounter["CancelBuild"])
	}
}
