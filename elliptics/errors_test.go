package elliptics

import (
	"errors"

	. "gopkg.in/check.v1"
)

func init() {
	Suite(&ErrorSuite{})
}

type ErrorSuite struct{}

func (s *ErrorSuite) TestDnetError(c *C) {
	var (
		dnetError = &DnetError{
			Code:    200,
			Flags:   1024,
			Message: "dummy dnet error",
		}

		dummyError = errors.New("dummy error")
	)

	c.Assert(DnetErrorFromError(dnetError), FitsTypeOf, &DnetError{})
	c.Assert(DnetErrorFromError(dummyError), IsNil)

	c.Assert(ErrorData(dnetError), Equals, dnetError.Message)
	c.Assert(ErrorData(dummyError), Equals, dummyError.Error())

	c.Assert(ErrorCode(dnetError), Equals, dnetError.Code)
	c.Assert(ErrorCode(dummyError), Equals, -22)
}
