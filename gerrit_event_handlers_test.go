package main

import "testing"

func TestShouldBeNoPatchsetCreatedHandlersByDefault(t *testing.T) {
	if len(eventRouter["patchset-created"]) != 0 {
		t.Error("Should be no patchset-created handlers by default")
	}

}
