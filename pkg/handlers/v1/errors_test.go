package v1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrMissingValue(t *testing.T) {
	field := "foo"
	e := ErrMissingValue{Field: field}
	require.Equal(t, fmt.Sprintf("no field %s was provided", field), e.Error())
}
