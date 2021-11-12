package py

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/FollowTheProcess/py/internal/test"
)

func TestGetPath(t *testing.T) {
	got, err := getPath()
	if err != nil {
		t.Fatalf("getPath returned an error: %v", err)
	}

	if len(got) == 0 {
		t.Fatal("length of returned path was 0")
	}

	// Since I don't like the idea of changing $PATH halfway through a test
	// and that just about every Unix system ever has a /usr/bin
	// this should be okay

	if !isIn("/usr/bin", got) {
		t.Error("/usr/bin not found in path")
	}
}

func TestIsPythonInterpreter(t *testing.T) {
	type args struct {
		path string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "normal python path",
			args: args{path: "/usr/local/bin/python"},
			want: true,
		},
		{
			name: "path with a version",
			args: args{path: "/usr/local/bin/python3.10"},
			want: true,
		},
		{
			name: "different stem",
			args: args{path: "/somewhere/else/python3.10"},
			want: true,
		},
		{
			name: "not python",
			args: args{path: "/usr/local/bin/dingle"},
			want: false,
		},
		{
			name: "not python with a version",
			args: args{path: "/usr/local/bin/dingle3.10"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPythonInterpreter(tt.args.path)
			if got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestGetPythonInterpreters(t *testing.T) {
	root, err := test.GetProjectRoot()
	if err != nil {
		t.Fatalf("could not get project root: %v", err)
	}
	testDir := filepath.Join(root, "testdata", "pythonpaths", "pythonpath1")

	type args struct {
		dir string
	}

	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "test",
			args: args{dir: testDir},
			want: []string{
				// The order of these matters
				filepath.Join(testDir, "python"),
				filepath.Join(testDir, "python3.10"),
				filepath.Join(testDir, "python3.9"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPythonInterpreters(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getPythonInterpreters() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if len(got) == 0 {
				t.Fatalf("length of returned paths was 0")
			}

			// If the 'notpython' file turns up
			if isIn("notpython", got) {
				t.Error("something other than a python interpreter was returned")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, wanted %#v", got, tt.want)
			}
		})
	}
}

func isIn(needle string, haystack []string) bool {
	for _, thing := range haystack {
		if thing == needle {
			return true
		}
	}
	return false
}
