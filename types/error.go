package types

import (
	"fmt"

	"github.com/pkg/errors"
)

// UnexpectedStatusCode returns an meaning error to be returned when getting unexpected status code.
func UnexpectedStatusCode(expected, got int) error {
	return fmt.Errorf("unexpected status code, expect %v, got %v", expected, got)
}

// ErrUnknownNetwork is returned when the given network is unknown to us.
var ErrUnknownNetwork = errors.New("unknown network")
