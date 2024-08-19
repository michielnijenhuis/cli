package helper

import (
	"fmt"
	"testing"
)

func TestEscapeSequencesCanBeStripped(t *testing.T) {
	tags := []string{"info", "error", "warn", "primary", "accent"}

	for _, tag := range tags {
		if StripEscapeSequences(fmt.Sprintf("<%s>foo</%s>", tag, tag)) != "foo" {
			t.Errorf("failed to strip tag \"%s\"", tag)
		}
	}

	if StripEscapeSequences("<fg=cyan>foo</>") != "foo" {
		t.Errorf("failed to strip \"fg=cyan\"")
	}

	if StripEscapeSequences("<fg=bright-white>foo</>") != "foo" {
		t.Errorf("failed to strip \"fg=cyan\"")
	}

	if StripEscapeSequences("<bg=cyan>foo</>") != "foo" {
		t.Errorf("failed to strip \"bg=cyan\"")
	}

	if StripEscapeSequences("<options=bold>foo</>") != "foo" {
		t.Errorf("failed to strip \"options=bold\"")
	}

	if StripEscapeSequences("<fg=cyan;bg=white;options=bold>foo</>") != "foo" {
		t.Errorf("failed to strip \"fg=cyan;bg=white;options=bold\"")
	}
}
