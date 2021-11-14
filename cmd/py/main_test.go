package main

import "testing"

func TestIsMajorSpecifier(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "valid 3",
			arg:  "-3",
			want: true,
		},
		{
			name: "valid 4",
			arg:  "-4",
			want: true,
		},
		{
			name: "valid 2",
			arg:  "-2",
			want: true,
		},
		{
			name: "no leading dash",
			arg:  "4",
			want: false,
		},
		{
			name: "no leading dash letter",
			arg:  "p",
			want: false,
		},
		{
			name: "exact version specifier",
			arg:  "-3.9",
			want: false,
		},
		{
			name: "more than 1 character after dash",
			arg:  "-27",
			want: false,
		},
		{
			name: "more than 1 character after dash letters",
			arg:  "-blah",
			want: false,
		},
		{
			name: "not a valid integer",
			arg:  "-p",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMajorSpecifier(tt.arg); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestIsExactSpecifier(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "valid 3.9",
			arg:  "-3.9",
			want: true,
		},
		{
			name: "valid 3.10",
			arg:  "-3.10",
			want: true,
		},
		{
			name: "valid 4.0",
			arg:  "-4.0",
			want: true,
		},
		{
			name: "major version specifier",
			arg:  "-3",
			want: false,
		},
		{
			name: "no leading dash",
			arg:  "3.9",
			want: false,
		},
		{
			name: "no leading dash",
			arg:  "3.9",
			want: false,
		},
		{
			name: "no dot",
			arg:  "-39",
			want: false,
		},
		{
			name: "whitespace on the left of the dot",
			arg:  "- .9",
			want: false,
		},
		{
			name: "whitespace on the right of the dot",
			arg:  "-3. ",
			want: false,
		},
		{
			name: "nothing on the left of the dot",
			arg:  "-.9",
			want: false,
		},
		{
			name: "nothing on the right of the dot",
			arg:  "-3.",
			want: false,
		},
		{
			name: "version includes patch",
			arg:  "-3.9.8",
			want: false,
		},
		{
			name: "major not valid int",
			arg:  "-blah.9",
			want: false,
		},
		{
			name: "minor not valid int",
			arg:  "-3.blah",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExactSpecifier(tt.arg); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}
