package urn

import (
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  URN
	}{
		{"simple", "myapp:users:2", URN{"myapp", "users", "2"}},
		{"id with colons", "myapp:tracks:a:b:c", URN{"myapp", "tracks", "a:b:c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			got, err := Parse(tt.input)
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.Equals, tt.want)
			c.Assert(got.String(), qt.Equals, tt.input)
		})
	}
}

func TestParseRejectsMalformed(t *testing.T) {
	for _, input := range []string{"", "users", "myapp:users", "myapp::2", ":users:2", "myapp:users:"} {
		t.Run(input, func(t *testing.T) {
			c := qt.New(t)
			_, err := Parse(input)
			c.Assert(errors.Is(err, ErrInvalid), qt.IsTrue)
		})
	}
}
