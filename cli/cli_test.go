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

func TestParsePyPython(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantMajor int
		wantMinor int
		wantErr   bool
	}{
		{
			name:      "valid 3.10",
			version:   "3.10",
			wantMajor: 3,
			wantMinor: 10,
			wantErr:   false,
		},
		{
			name:      "valid 3.9",
			version:   "3.9",
			wantMajor: 3,
			wantMinor: 9,
			wantErr:   false,
		},
		{
			name:      "valid 4.0",
			version:   "4.0",
			wantMajor: 4,
			wantMinor: 0,
			wantErr:   false,
		},
		{
			name:      "no major",
			version:   ".9",
			wantMajor: 0,
			wantMinor: 0,
			wantErr:   true,
		},
		{
			name:      "no minor",
			version:   "3.",
			wantMajor: 0,
			wantMinor: 0,
			wantErr:   true,
		},
		{
			name:      "no dot",
			version:   "39",
			wantMajor: 0,
			wantMinor: 0,
			wantErr:   true,
		},
		{
			name:      "only one number",
			version:   "3",
			wantMajor: 0,
			wantMinor: 0,
			wantErr:   true,
		},
		{
			name:      "empty string",
			version:   "",
			wantMajor: 0,
			wantMinor: 0,
			wantErr:   true,
		},
		{
			name:      "whitespace before",
			version:   " 3.9",
			wantMajor: 0,
			wantMinor: 0,
			wantErr:   true,
		},
		{
			name:      "whitespace after",
			version:   "3.9 ",
			wantMajor: 0,
			wantMinor: 0,
			wantErr:   true,
		},
		{
			name:      "major not an int",
			version:   "X.9",
			wantMajor: 0,
			wantMinor: 0,
			wantErr:   true,
		},
		{
			name:      "minor not an int",
			version:   "3.X",
			wantMajor: 0,
			wantMinor: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMajor, gotMinor, err := parsePyPython(tt.version)

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePyPython() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if gotMajor != tt.wantMajor {
				t.Errorf("wrong major version. got %d, wanted %d", gotMajor, tt.wantMajor)
			}

			if gotMinor != tt.wantMinor {
				t.Errorf("wrong minor version. got %d, wanted %d", gotMinor, tt.wantMinor)
			}
		})
	}
}
