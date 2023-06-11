package cli //nolint: testpackage // Need access to internals

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

// newTestApp creates and returns a test App object configured to talk to 'out' and 'err'
// with a mocked out $PATH, given by 'path'.
func newTestApp(out, err io.Writer, path string) *App {
	return &App{
		Stdout: out,
		Stderr: err,
		Path:   path,
		Logger: logrus.New(), // Doesn't actually matter but it needs it to work
	}
}

func TestApp_Help(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	path := "" // Doesn't matter for help

	app := newTestApp(stdout, stderr, path)

	want := fmt.Sprintf("%s\n", helpText)
	app.Help()

	if got := stdout.String(); got != want {
		t.Errorf("got %#v, wanted %#v", got, want)
	}
}

func TestApp_getPathEntries(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		path    string
		want    []string
		wantErr bool
	}{
		{
			name: "normal path",
			path: "/usr/bin:/usr/local/bin:/usr/local/somewhere",
			want: []string{"/usr/bin", "/usr/local/bin", "/usr/local/somewhere"},
		},
		{
			name: "empty",
			path: "",
			want: []string{},
		},
		{
			name: "duplicate entries",
			path: "/usr/bin:/usr/local/bin:/usr/bin:/usr/somewhere:/usr/local/bin",
			want: []string{"/usr/bin", "/usr/local/bin", "/usr/somewhere"},
		},
		{
			name: "empty entry should be replaced with .",
			path: "/usr/bin:/usr/local/bin::/usr/somewhere:",
			want: []string{"/usr/bin", "/usr/local/bin", ".", "/usr/somewhere"},
		},
		{
			name: "multiple empty entries should be one .",
			path: "/usr/bin::/usr/local/bin::/usr/somewhere:",
			want: []string{"/usr/bin", ".", "/usr/local/bin", "/usr/somewhere"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make the test App
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			app := newTestApp(stdout, stderr, tt.path)

			// Get the value using the key specified in the test case
			got := app.getPathEntries()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, wanted %#v", got, tt.want)
			}
		})
	}
}

func Test_exists(t *testing.T) {
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

func Test_parsePyPython(t *testing.T) {
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
			app := &App{}
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

func Test_parseShebang(t *testing.T) {
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
			app := &App{Stdout: os.Stdout, Logger: logrus.New()}
			if got := app.parseShebang(tt.shebang); got != tt.want {
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}

func Test_deDupe(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "only 1 path",
			args: args{paths: []string{"its/just/me/here"}},
			want: []string{"its/just/me/here"},
		},
		{
			name: "3 different paths",
			args: args{paths: []string{"a/path", "another/path", "athird/path"}},
			want: []string{"a/path", "another/path", "athird/path"},
		},
		{
			name: "1 unique, 2 duplicates",
			args: args{paths: []string{"a/path", "a/path", "aunique/path"}},
			want: []string{"a/path", "aunique/path"},
		},
		{
			name: "all duplicates",
			args: args{paths: []string{"a/path", "a/path", "a/path"}},
			want: []string{"a/path"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deDupe(tt.args.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deDupe() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Not really necessary but I was just curious and it was easy to do.
func Benchmark_deDupe(b *testing.B) {
	// Some paths that contain duplicates
	paths := []string{
		"/usr/local/bin/python3.7",
		"/usr/local/bin/python3.8",
		"/usr/local/bin/python3.9",
		"/usr/local/bin/python3.6",
		"/usr/local/bin/python3.10",
		"/usr/local/bin/python3.11",
		"/usr/bin/python3.7",
		"/usr/bin/python3.8",
		"/usr/bin/python2.7",
		"/usr/bin/python",
		"/usr/bin/python2",
		"/usr/bin/python3",
		"/usr/bin/python2",
		"/usr/bin/python",
		"/usr/bin/python",
		"/usr/bin/python",
		"/usr/bin/python3",
		"/usr/local/bin/python3.11",
		"/usr/local/bin/python3.9",
		"/usr/local/bin/python3.9",
		"/usr/local/bin/python3.6",
	}

	// Reset prior to actually running the benchmark
	// ensures we don't include the initialisation stuff
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		deDupe(paths)
	}
}
