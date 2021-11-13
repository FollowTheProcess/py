package py

import (
	"reflect"
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
