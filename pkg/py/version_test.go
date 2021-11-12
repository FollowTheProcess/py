package py

import (
	"reflect"
	"testing"
)

func TestVersion_FromFileName(t *testing.T) {
	type args struct {
		filename string
	}

	tests := []struct {
		name    string
		args    args
		want    Version
		wantErr bool
	}{
		{
			name:    "valid 3.7",
			args:    args{filename: "python3.7"},
			want:    Version{Major: 3, Minor: 7},
			wantErr: false,
		},
		{
			name:    "valid 3.10",
			args:    args{filename: "python3.10"},
			want:    Version{Major: 3, Minor: 10},
			wantErr: false,
		},
		{
			name:    "valid 4.0",
			args:    args{filename: "python4.0"},
			want:    Version{Major: 4, Minor: 0},
			wantErr: false,
		},
		{
			name:    "valid 2.7",
			args:    args{filename: "python2.7"},
			want:    Version{Major: 2, Minor: 7},
			wantErr: false,
		},
		{
			name:    "no version numbers",
			args:    args{filename: "python"},
			want:    Version{},
			wantErr: true,
		},
		{
			name:    "just the major version (should never happen)",
			args:    args{filename: "python3"},
			want:    Version{},
			wantErr: true,
		},
		{
			name:    "filename not python",
			args:    args{filename: "dingle3.7"},
			want:    Version{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{}
			err := v.FromFileName(tt.args.filename)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FromFileName() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(v, tt.want) {
				t.Errorf("got %#v, wanted %#v", v, tt.want)
			}
		})
	}
}
