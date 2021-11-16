package interpreter

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"
)

func TestInterpreter_FromFilePath(t *testing.T) {
	type args struct {
		filepath string
	}

	tests := []struct {
		name    string
		args    args
		want    Interpreter
		wantErr bool
	}{
		{
			name:    "valid 3.7",
			args:    args{filepath: "/usr/local/bin/python3.7"},
			want:    Interpreter{Major: 3, Minor: 7, Path: "/usr/local/bin/python3.7"},
			wantErr: false,
		},
		{
			name:    "valid 3.10",
			args:    args{filepath: "/usr/local/bin/python3.10"},
			want:    Interpreter{Major: 3, Minor: 10, Path: "/usr/local/bin/python3.10"},
			wantErr: false,
		},
		{
			name:    "valid 4.0",
			args:    args{filepath: "/usr/local/bin/python4.0"},
			want:    Interpreter{Major: 4, Minor: 0, Path: "/usr/local/bin/python4.0"},
			wantErr: false,
		},
		{
			name:    "valid 2.7",
			args:    args{filepath: "/usr/bin/python2.7"},
			want:    Interpreter{Major: 2, Minor: 7, Path: "/usr/bin/python2.7"},
			wantErr: false,
		},
		{
			name:    "no Interpreter numbers (usually means system python)",
			args:    args{filepath: "/usr/bin/python"},
			want:    Interpreter{},
			wantErr: true,
		},
		{
			name:    "just the major Interpreter (again usually system python)",
			args:    args{filepath: "/usr/bin/python3"},
			want:    Interpreter{},
			wantErr: true,
		},
		{
			name:    "filepath not python",
			args:    args{filepath: "/usr/local/bin/dingle3.7"},
			want:    Interpreter{},
			wantErr: true,
		},
		{
			name:    "bad major component",
			args:    args{filepath: "/usr/local/bin/pythonp.7"},
			want:    Interpreter{},
			wantErr: true,
		},
		{
			name:    "bad minor component",
			args:    args{filepath: "/usr/local/bin/python3.f"},
			want:    Interpreter{},
			wantErr: true,
		},
		{
			name:    "whitespace before",
			args:    args{filepath: "/usr/local/bin python 3.7"},
			want:    Interpreter{},
			wantErr: true,
		},
		{
			name:    "whitespace between python and Interpreter",
			args:    args{filepath: "/usr/local/bin/python 3.7"},
			want:    Interpreter{},
			wantErr: true,
		},
		{
			name:    "whitespace after",
			args:    args{filepath: "/usr/local/bin/python3.7 "},
			want:    Interpreter{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Interpreter{}
			err := v.FromFilePath(tt.args.filepath)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Fromfilepath() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(v, tt.want) {
				t.Errorf("got %#v, wanted %#v", v, tt.want)
			}
		})
	}
}

func TestInterpreter_ToString(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Path  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "python 3.10",
			fields: fields{Major: 3, Minor: 10, Path: "/usr/local/bin/python3.10"},
			want:   "3.10\t│ /usr/local/bin/python3.10",
		},
		{
			name:   "python 3.9",
			fields: fields{Major: 3, Minor: 9, Path: "/usr/local/bin/python3.9"},
			want:   "3.9\t│ /usr/local/bin/python3.9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interpreter{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Path:  tt.fields.Path,
			}
			if got := i.ToString(); got != tt.want {
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}

func TestInterpreter_String(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Path  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "python 3.10",
			fields: fields{Major: 3, Minor: 10, Path: "/usr/local/bin/python3.10"},
			want:   "/usr/local/bin/python3.10",
		},
		{
			name:   "python 3.9",
			fields: fields{Major: 3, Minor: 9, Path: "/usr/local/bin/python3.9"},
			want:   "/usr/local/bin/python3.9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interpreter{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Path:  tt.fields.Path,
			}
			if got := i.String(); got != tt.want {
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}

func TestListSort(t *testing.T) {
	tests := []struct {
		name string
		list List
		want List
	}{
		{
			name: "test",
			list: List{
				// We don't intialise paths because it doesn't matter for sorting
				{
					Major: 3,
					Minor: 7,
				},
				{
					Major: 3,
					Minor: 4,
				},
				{
					Major: 2,
					Minor: 7,
				},
				{
					Major: 3,
					Minor: 10,
				},
				{
					Major: 4,
					Minor: 1,
				},
				{
					Major: 3,
					Minor: 11,
				},
				{
					Major: 4,
					Minor: 0,
				},
				{
					Major: 3,
					Minor: 12,
				},
				{
					Major: 4,
					Minor: 10,
				},
				{
					Major: 2,
					Minor: 6,
				},
			},
			want: List{
				{
					Major: 4,
					Minor: 10,
				},
				{
					Major: 4,
					Minor: 1,
				},
				{
					Major: 4,
					Minor: 0,
				},
				{
					Major: 3,
					Minor: 12,
				},
				{
					Major: 3,
					Minor: 11,
				},
				{
					Major: 3,
					Minor: 10,
				},
				{
					Major: 3,
					Minor: 7,
				},
				{
					Major: 3,
					Minor: 4,
				},
				{
					Major: 2,
					Minor: 7,
				},
				{
					Major: 2,
					Minor: 6,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.list) != len(tt.want) {
				t.Fatalf("len(list): %d, len(want): %d. Check the test cases", len(tt.list), len(tt.want))
			}
			sort.Sort(tt.list)
			// Now tt.list should be sorted and match tt.want
			if !reflect.DeepEqual(tt.list, tt.want) {
				t.Errorf("got %v, wanted %v", tt.list, tt.want)
			}
		})
	}
}

func TestInterpreter_SatisfiesMajor(t *testing.T) {
	type args struct {
		version int
	}

	tests := []struct {
		name        string
		interpreter Interpreter
		args        args
		want        bool
	}{
		{
			name:        "3.8 satisfies 3",
			interpreter: Interpreter{Major: 3, Minor: 8},
			args:        args{version: 3},
			want:        true,
		},
		{
			name:        "3.9 satisfies 3",
			interpreter: Interpreter{Major: 3, Minor: 9},
			args:        args{version: 3},
			want:        true,
		},
		{
			name:        "3.10 satisfies 3",
			interpreter: Interpreter{Major: 3, Minor: 10},
			args:        args{version: 3},
			want:        true,
		},
		{
			name:        "2.7 does not satisfy 3",
			interpreter: Interpreter{Major: 2, Minor: 7},
			args:        args{version: 3},
			want:        false,
		},
		{
			name:        "4.1 does not satisfy 3",
			interpreter: Interpreter{Major: 4, Minor: 1},
			args:        args{version: 3},
			want:        false,
		},
		{
			name:        "4.0 satisfies 4",
			interpreter: Interpreter{Major: 4, Minor: 0},
			args:        args{version: 4},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interpreter.SatisfiesMajor(tt.args.version); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestInterpreter_SatisfiesExact(t *testing.T) {
	type args struct {
		major int
		minor int
	}

	tests := []struct {
		name        string
		interpreter Interpreter
		args        args
		want        bool
	}{
		{
			name:        "3.9 satisfies 3.9",
			interpreter: Interpreter{Major: 3, Minor: 9},
			args:        args{major: 3, minor: 9},
			want:        true,
		},
		{
			name:        "3.8 satisfies 3.8",
			interpreter: Interpreter{Major: 3, Minor: 8},
			args:        args{major: 3, minor: 8},
			want:        true,
		},
		{
			name:        "3.7 satisfies 3.7",
			interpreter: Interpreter{Major: 3, Minor: 7},
			args:        args{major: 3, minor: 7},
			want:        true,
		},
		{
			name:        "9.12 satisfies 9.12",
			interpreter: Interpreter{Major: 9, Minor: 12},
			args:        args{major: 9, minor: 12},
			want:        true,
		},
		{
			name:        "2.7 does not satisfy 3.7",
			interpreter: Interpreter{Major: 2, Minor: 7},
			args:        args{major: 3, minor: 7},
			want:        false,
		},
		{
			name:        "3.7 does not satisfy 2.7",
			interpreter: Interpreter{Major: 3, Minor: 7},
			args:        args{major: 2, minor: 7},
			want:        false,
		},
		{
			name:        "3.10 does not satisfy 3.11",
			interpreter: Interpreter{Major: 3, Minor: 10},
			args:        args{major: 3, minor: 11},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interpreter.SatisfiesExact(tt.args.major, tt.args.minor); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func Test_getPythonInterpreters(t *testing.T) {
	root, err := getProjectRoot()
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
		want    List
		wantErr bool
	}{
		{
			name: "test",
			args: args{dir: testDir},
			want: List{
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
	root, err := getProjectRoot()
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
		want    List
		wantErr bool
	}{
		{
			name: "test",
			args: args{paths: []string{
				filepath.Join(testDir, "pythonpath1"),
				filepath.Join(testDir, "pythonpath2"),
				filepath.Join(testDir, "pythonpath3"),
			}},
			want: List{
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
					Major: 3,
					Minor: 5,
					Path:  filepath.Join(testDir, "pythonpath3", "python3.5"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAll(tt.args.paths)
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

func BenchmarkGetAllPythonInterpreters(b *testing.B) {
	root, err := getProjectRoot()
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
		if _, err := GetAll(paths); err != nil {
			b.Fatalf("GetAllPythonInterpreters returned an error during benchmarking: %s", err)
		}
	}
}

func BenchmarkInterpreterSort(b *testing.B) {
	input := List(List{
		{
			Major: 3,
			Minor: 7,
		},
		{
			Major: 3,
			Minor: 4,
		},
		{
			Major: 2,
			Minor: 7,
		},
		{
			Major: 3,
			Minor: 10,
		},
		{
			Major: 4,
			Minor: 1,
		},
		{
			Major: 3,
			Minor: 11,
		},
		{
			Major: 4,
			Minor: 0,
		},
		{
			Major: 3,
			Minor: 12,
		},
		{
			Major: 4,
			Minor: 10,
		},
		{
			Major: 2,
			Minor: 6,
		},
	})

	// Reset prior to actually running the benchmark
	// ensures we don't include the initialisation stuff
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		sort.Sort(input)
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

// getProjectRoot is a convenience function for reliably getting the project root dir from anywhere
// so that tests can make use of root-relative paths
func getProjectRoot() (string, error) {
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not find current filepath")
	}

	return filepath.Join(filepath.Dir(here), "../.."), nil
}
