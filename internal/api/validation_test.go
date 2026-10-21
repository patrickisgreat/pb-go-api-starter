package api

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestIsValidUrn(t *testing.T) {
	tests := []struct {
		name  string
		urn   string
		types []string
		want  bool
	}{
		{"users urn", "myapp:users:2", []string{"users"}, true},
		{"any collection when unrestricted", "myapp:tracks:2", nil, true},
		{"wrong collection", "myapp:tracks:2", []string{"users"}, false},
		{"empty", "", []string{"users"}, false},
		{"malformed", "users:2", []string{"users"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(isValidUrn(tt.urn, tt.types), qt.Equals, tt.want)
		})
	}
}
