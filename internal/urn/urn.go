// Package urn parses the identifiers used in request metadata.
//
// A URN has the form "namespace:collection:id", for example "myapp:users:42".
// The collection names the kind of thing being identified and is what callers
// usually validate against.
package urn

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalid is returned by Parse when the input is not a well formed URN.
var ErrInvalid = errors.New("invalid urn")

// URN is a parsed identifier.
type URN struct {
	Namespace  string
	Collection string
	ID         string
}

// Parse splits s into its three parts. Every part must be non-empty; the id
// may itself contain colons.
func Parse(s string) (URN, error) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) != 3 {
		return URN{}, fmt.Errorf("%w: %q must have the form namespace:collection:id", ErrInvalid, s)
	}
	for _, p := range parts {
		if p == "" {
			return URN{}, fmt.Errorf("%w: %q has an empty part", ErrInvalid, s)
		}
	}
	return URN{Namespace: parts[0], Collection: parts[1], ID: parts[2]}, nil
}

// String renders the URN back to its textual form.
func (u URN) String() string {
	return u.Namespace + ":" + u.Collection + ":" + u.ID
}
