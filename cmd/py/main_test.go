package main

import (
	"bytes"
	"testing"

	"github.com/FollowTheProcess/py/cli"
)

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

func TestParseMajorSpecifier(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want int
	}{
		{
			name: "major 3",
			arg:  "-3",
			want: 3,
		},
		{
			name: "major 2",
			arg:  "-2",
			want: 2,
		},
		{
			name: "major 4",
			arg:  "-4",
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseMajorSpecifier(tt.arg); got != tt.want {
				t.Errorf("got %d, wanted %d", got, tt.want)
			}
		})
	}
}

func TestParseExactSpecifier(t *testing.T) {
	tests := []struct {
		name      string
		arg       string
		wantMajor int
		wantMinor int
	}{
		{
			name:      "3.9",
			arg:       "-3.9",
			wantMajor: 3,
			wantMinor: 9,
		},
		{
			name:      "2.7",
			arg:       "-2.7",
			wantMajor: 2,
			wantMinor: 7,
		},
		{
			name:      "3.10",
			arg:       "-3.10",
			wantMajor: 3,
			wantMinor: 10,
		},
		{
			name:      "4.0",
			arg:       "-4.0",
			wantMajor: 4,
			wantMinor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor := parseExactSpecifier(tt.arg)
			if major != tt.wantMajor {
				t.Errorf("major version difference, got %d, wanted %d", major, tt.wantMajor)
			}
			if minor != tt.wantMinor {
				t.Errorf("minor version difference, got %d, wanted %d", minor, tt.wantMinor)
			}
		})
	}
}

func TestCLIFlags(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		args    []string
		wantErr bool
	}{
		// Take care not to do anything here that starts a REPL
		{
			name:    "--list with extra arg",
			args:    []string{"--list", "something"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "--help",
			args:    []string{"--help"},
			want:    "",
			wantErr: false,
		},
		{
			name:    "--help with extra arg",
			args:    []string{"--help", "something"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appOut := &bytes.Buffer{}
			appErr := &bytes.Buffer{}

			app := cli.New(appOut, appErr)

			err := run(app, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
