package util

import "encoding/json"

// Optional represents a JSON field that can be absent, present with a value, or present as null.
// Present is set to true only when the field appears in the JSON payload.
type Optional[T any] struct {
	Value   T
	Present bool
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Present = true
	return json.Unmarshal(data, &o.Value)
}

// NewOptional creates a present Optional with the given value. Useful in tests.
func NewOptional[T any](v T) Optional[T] {
	return Optional[T]{Value: v, Present: true}
}
