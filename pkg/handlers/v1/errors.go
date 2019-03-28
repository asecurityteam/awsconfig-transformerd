package v1

import "fmt"

// ErrMissingValue is returned when a configuration item is missing a required field
type ErrMissingValue struct {
	Field string
}

func (e ErrMissingValue) Error() string {
	return fmt.Sprintf("no field %s was provided", e.Field)
}
