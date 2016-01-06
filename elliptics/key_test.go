package elliptics

import (
	"crypto/sha512"
	"errors"
	. "gopkg.in/check.v1"
)

type KeySuite struct {
	badKey  int
	goodKey string
}

func init() {
	Suite(&KeySuite{
		badKey:  9999,
		goodKey: "some_key",
	})

	Suite(&DnetErrorSuite{})
}

func (s *KeySuite) TestKeyDefaultCreationAndFree(c *C) {
	key, err := NewKey()
	c.Assert(err, IsNil)
	defer key.Free()

	c.Assert(key.ById(), Equals, false)
}

func (s *KeySuite) TestKeyCreationAndFree(c *C) {
	_, err := NewKey(s.badKey)
	c.Assert(err, Equals, InvalidKeyArgument)

	key, err := NewKey(s.goodKey)
	c.Assert(err, IsNil)
	key.Free()
}

func (s *KeySuite) TestKeysCreationAndFree(c *C) {
	c.Skip("Skip this test")
	keys, err := NewKeys([]string{"A", "B", "C"})
	c.Assert(err, IsNil)
	defer keys.Free()

	var hash []uint8
	for _, v := range sha512.Sum512([]byte("A")) {
		hash = append(hash, v)
	}
	name, err := keys.Find(hash)
	if err != nil {
		c.Errorf("Find: Unexpected error %s", err)
	}

	if name != "A" {
		c.Errorf("Unexpected `name` value %s", name)
	}
}

type DnetErrorSuite struct{}

func (s *KeySuite) TestDnetError(c *C) {
	var (
		dnetCode = 100
		dnetFlag = uint64(16)
		dnetMsg  = "dummy_dnet_error_message"

		derr = &DnetError{
			Code:    dnetCode,
			Flags:   dnetFlag,
			Message: dnetMsg,
		}

		dummyErr = errors.New("dummy_err")
	)

	c.Assert(ErrorData(derr), Equals, dnetMsg)
	c.Assert(ErrorData(dummyErr), Equals, dummyErr.Error())
	c.Assert(derr.Error(), Not(HasLen), 0)
}
