package main

import (
	"testing"
)

func TestItCreatesABuildWhenTheCommentIsRetest(t *testing.T) {
	p := NewMockPipeline()
	b := NewMockBackend()
	event := Event{
		PatchSet: PatchSet{
			Number:   1,
			Revision: "123456",
		},
		Change: Change{
			Number: 1,
		},
		Comment: "retest",
	}

	// When it's just "retest"
	HandleCommentAdded(event, p, b)
	if p.FunctionCallCounter["CreateBuild"] != 1 {
		t.Errorf("Expected CreateBuild to be called once, but it was called %d times", p.FunctionCallCounter["CreateBuild"])
	}

	p.Reset("CreateBuild")
	// When retest is on a line by itself
	event.Comment = `
retest
		`
	HandleCommentAdded(event, p, b)
	if p.FunctionCallCounter["CreateBuild"] != 1 {
		t.Errorf("Expected CreateBuild to be called once, but it was called %d times", p.FunctionCallCounter["CreateBuild"])
	}

	p.Reset("CreateBuild")
	// When retest is on a line by itself in a larger comment
	event.Comment = `
Just to make sure
retest
When this is done the change can be merged
		`
	HandleCommentAdded(event, p, b)
	if p.FunctionCallCounter["CreateBuild"] != 1 {
		t.Errorf("Expected CreateBuild to be called once, but it was called %d times", p.FunctionCallCounter["CreateBuild"])
	}

	// Not when retest is within a line
	p.Reset("CreateBuild")
	event.Comment = "not retest"
	HandleCommentAdded(event, p, b)
	if c, ok := p.FunctionCallCounter["CreateBuild"]; ok && c != 0 {
		t.Errorf("Expected CreateBuild to be called zero times, but it was called %d times", p.FunctionCallCounter["CreateBuild"])
	}

	// Not when retest has leading whitespace
	event.Comment = " retest"
	HandleCommentAdded(event, p, b)
	if c, ok := p.FunctionCallCounter["CreateBuild"]; ok && c != 0 {
		t.Errorf("Expected CreateBuild to be called zero times, but it was called %d times", p.FunctionCallCounter["CreateBuild"])
	}
	// Not when retest has trailing whitespace
	event.Comment = "retest "
	HandleCommentAdded(event, p, b)
	if c, ok := p.FunctionCallCounter["CreateBuild"]; ok && c != 0 {
		t.Errorf("Expected CreateBuild to be called zero times, but it was called %d times", p.FunctionCallCounter["CreateBuild"])
	}
	// Note the leading whitespace on its own line
	event.Comment = `
	retest
	`
	HandleCommentAdded(event, p, b)
	if c, ok := p.FunctionCallCounter["CreateBuild"]; ok && c != 0 {
		t.Errorf("Expected CreateBuild to be called zero times, but it was called %d times", p.FunctionCallCounter["CreateBuild"])
	}
}
