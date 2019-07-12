package types

import (
	"fmt"

	"github.com/pkg/errors"
)

type ErrHTTPResponse struct {
	error

	ExpectedStatusCode int
	GotStatusCode      int
	GotBody            []byte
}

func NewErrHTTPResponse(expectedStatusCode, gotStatusCode int, gotBody []byte) error {
	return ErrHTTPResponse{
		error:              fmt.Errorf("bad http response: expected status=%v got status=%v\nbody: %v", expectedStatusCode, gotStatusCode, string(gotBody)),
		ExpectedStatusCode: expectedStatusCode,
		GotStatusCode:      gotStatusCode,
		GotBody:            gotBody,
	}
}

// UnexpectedStatusCode returns an meaning error to be returned when getting unexpected status code.
func UnexpectedStatusCode(expected, got int) error {
	return fmt.Errorf("unexpected status code, expect %v, got %v", expected, got)
}

// ErrList is a list of errors.
type ErrList []error

func NewErrList(n int) ErrList {
	return make([]error, n)
}

// ErrList implements the error interface.
func (errs ErrList) Error() string {
	errMsg := ""
	for i := range errs {
		errMsg += fmt.Sprintf("[%v] %v, ", i, errs[i].Error())
	}
	return errMsg
}

// ErrUnknownNetwork is returned when the given network is unknown to us.
var ErrUnknownNetwork = errors.New("unknown network")
