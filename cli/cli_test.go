package cli

import (
	"bytes"
	"os"
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

func TestExists(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get cwd")
	}
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "cwd must (presumably) exist",
			path: cwd,
			want: true,
		},
		{
			name: "something made up",
			path: "im/not/here",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exists(tt.path); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
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
			app := New()
			gotMajor, gotMinor, err := app.parsePyPython(tt.version)

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

func TestParseShebang(t *testing.T) {
	tests := []struct {
		name    string
		shebang string
		want    string
	}{
		{
			name:    "python3 returns 3",
			shebang: "#!/usr/bin/python3",
			want:    "3",
		},
		{
			name:    "python3.9 returns 3.9",
			shebang: "#!/usr/bin/python3.9",
			want:    "3.9",
		},
		{
			name:    "python3.10 returns 3.10",
			shebang: "#!/usr/bin/python3.10",
			want:    "3.10",
		},
		{
			name:    "python4.12 returns 4.12",
			shebang: "#!/usr/bin/python4.12",
			want:    "4.12",
		},
		{
			name:    "no version returns nothing",
			shebang: "#!/usr/bin/python",
			want:    "",
		},
		{
			name:    "no version returns nothing (local)",
			shebang: "#!/usr/local/bin/python",
			want:    "",
		},
		{
			name:    "no version returns nothing (env)",
			shebang: "#!/usr/bin/env python",
			want:    "",
		},
		{
			name:    "local 3.9 returns 3.9",
			shebang: "#!/usr/local/bin/python3.9",
			want:    "3.9",
		},
		{
			name:    "env 3.9 returns 3.9",
			shebang: "#!/usr/bin/env python3.9",
			want:    "3.9",
		},
		{
			name:    "local 4.12 returns 4.12",
			shebang: "#!/usr/local/bin/python4.12",
			want:    "4.12",
		},
		{
			name:    "env 4.12 returns 4.12",
			shebang: "#!/usr/bin/env python4.12",
			want:    "4.12",
		},
		{
			name:    "whitespace isn't counted",
			shebang: "#! /usr/bin/env python3",
			want:    "3",
		},
		{
			name:    "no #! means no shebang",
			shebang: "/usr/bin/python",
			want:    "",
		},
		{
			name:    "non valid path returns nothing",
			shebang: "#!/somewhere/not/recognised/python",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New()
			if got := app.parseShebang(tt.shebang); got != tt.want {
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}
