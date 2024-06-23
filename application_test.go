package cli

import (
	"testing"
)

func TestApplicationName(t *testing.T) {
	app := NewApplication("test", "v1.0.0")
	expected := "test"
	if got := app.GetName(); got != expected {
		t.Errorf("Got %s, want %s", got, expected)
	}
}