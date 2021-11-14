package cli

import (
	"bytes"
	"testing"
)

func TestAppVersion(t *testing.T) {
	out := &bytes.Buffer{}
	app := &App{Out: out}

	want := "py version: dev\ncommit: \n"
	app.Version()

	if got := out.String(); got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
