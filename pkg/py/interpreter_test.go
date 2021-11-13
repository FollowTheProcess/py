package py

import (
	"reflect"
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
				t.Fatalf("Fromfilepath() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(v, tt.want) {
				t.Errorf("got %#v, wanted %#v", v, tt.want)
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
			if got := i.String(); got != tt.want {
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}

func TestInterpreterListSort(t *testing.T) {
	tests := []struct {
		name string
		list InterpreterList
		want InterpreterList
	}{
		{
			name: "test",
			list: InterpreterList{
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
			want: InterpreterList{
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

func BenchmarkInterpreterSort(b *testing.B) {
	input := InterpreterList(InterpreterList{
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