package py

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/FollowTheProcess/py/internal/test"
)

func Test_getPythonInterpreters(t *testing.T) {
	root, err := test.GetProjectRoot()
	if err != nil {
		t.Fatalf("could not get project root: %s", err)
	}
	testDir := filepath.Join(root, "testdata", "pythonpaths", "pythonpath1")

	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    InterpreterList
		wantErr bool
	}{
		{
			name: "test",
			args: args{dir: testDir},
			want: InterpreterList{
				{
					Major: 3,
					Minor: 10,
					Path:  filepath.Join(testDir, "python3.10"),
				},
				{
					Major: 3,
					Minor: 9,
					Path:  filepath.Join(testDir, "python3.9"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPythonInterpreters(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPythonInterpreters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, wanted %#v", got, tt.want)
			}
		})
	}
}

func TestGetAllPythonInterpreters(t *testing.T) {
	root, err := test.GetProjectRoot()
	if err != nil {
		t.Fatalf("could not get project root: %s", err)
	}
	testDir := filepath.Join(root, "testdata", "pythonpaths")
	type args struct {
		paths []string
	}
	tests := []struct {
		name    string
		args    args
		want    InterpreterList
		wantErr bool
	}{
		{
			name: "test",
			args: args{paths: []string{
				filepath.Join(testDir, "pythonpath1"),
				filepath.Join(testDir, "pythonpath2"),
				filepath.Join(testDir, "pythonpath3"),
			}},
			want: InterpreterList{
				{
					Major: 3,
					Minor: 10,
					Path:  filepath.Join(testDir, "pythonpath1", "python3.10"),
				},
				{
					Major: 3,
					Minor: 9,
					Path:  filepath.Join(testDir, "pythonpath1", "python3.9"),
				},
				{
					Major: 3,
					Minor: 7,
					Path:  filepath.Join(testDir, "pythonpath2", "python3.7"),
				},
				{
					Major: 3,
					Minor: 8,
					Path:  filepath.Join(testDir, "pythonpath2", "python3.8"),
				},
				{
					Major: 2,
					Minor: 7,
					Path:  filepath.Join(testDir, "pythonpath3", "python2.7"),
				},
				{
					Major: 3,
					Minor: 5,
					Path:  filepath.Join(testDir, "pythonpath3", "python3.5"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllPythonInterpreters(tt.args.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllPythonInterpreters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestGetPath(t *testing.T) {
	// The default key for the test, so we can look for a wrong key
	// get a missing env var and test the error handling
	defaultKey := "PY_GETPATH_TEST"

	tests := []struct {
		name    string
		key     string
		path    string
		want    []string
		wantErr bool
	}{
		{
			name:    "normal path",
			key:     defaultKey,
			path:    "/usr/bin:/usr/local/bin:/usr/local/somewhere",
			want:    []string{"/usr/bin", "/usr/local/bin", "/usr/local/somewhere"},
			wantErr: false,
		},
		{
			name:    "empty",
			key:     defaultKey,
			path:    "",
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "duplicate entries",
			key:     defaultKey,
			path:    "/usr/bin:/usr/local/bin:/usr/bin:/usr/somewhere:/usr/local/bin",
			want:    []string{"/usr/bin", "/usr/local/bin", "/usr/somewhere"},
			wantErr: false,
		},
		{
			name:    "empty entry should be replaced with .",
			key:     defaultKey,
			path:    "/usr/bin:/usr/local/bin::/usr/somewhere:",
			want:    []string{"/usr/bin", "/usr/local/bin", ".", "/usr/somewhere"},
			wantErr: false,
		},
		{
			name:    "multiple empty entries should be one .",
			key:     defaultKey,
			path:    "/usr/bin::/usr/local/bin::/usr/somewhere:",
			want:    []string{"/usr/bin", ".", "/usr/local/bin", "/usr/somewhere"},
			wantErr: false,
		},
		{
			name:    "unset key",
			key:     "SOMETHING_ELSE",
			path:    "/usr/bin:/usr/local/bin:/usr/local/somewhere",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set our fake $PATH to the defaultKey
			t.Setenv(defaultKey, tt.path)

			// Get the value using the key specified in the test case
			got, err := GetPath(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPath() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, wanted %#v", got, tt.want)
			}
		})
	}
}

func BenchmarkGetAllPythonInterpreters(b *testing.B) {
	root, err := test.GetProjectRoot()
	if err != nil {
		b.Fatalf("could not get project root: %s", err)
	}
	testDir := filepath.Join(root, "testdata", "pythonpaths")

	paths := []string{
		filepath.Join(testDir, "pythonpath1"),
		filepath.Join(testDir, "pythonpath2"),
		filepath.Join(testDir, "pythonpath3"),
	}

	// Reset prior to actually running the benchmark
	// ensures we don't include the initialisation stuff
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if _, err := GetAllPythonInterpreters(paths); err != nil {
			b.Fatalf("GetAllPythonInterpreters returned an error during benchmarking: %s", err)
		}
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

// Not really necessary but I was just curious and it was easy to do
func BenchmarkDeDupe(b *testing.B) {
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
