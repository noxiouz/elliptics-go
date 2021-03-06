package elliptics

import (
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

func (s *KeySuite) TestKeyDefaultNewAndFree(c *C) {
	key, err := NewKey()
	c.Assert(err, IsNil)
	defer key.Free()

	c.Assert(key.ById(), Equals, false)
}

func (s *KeySuite) TestKeySetIdSetRawId(c *C) {
	key, err := NewKey()
	c.Assert(err, IsNil)
	defer key.Free()

	var id = []byte("21b4f4bd9e64ed355c3eb676a28ebedaf6d8f17bdc365995b319097153044080516bd083bfcce66121a3072646994c8430cc382b8dc543e84880183bf856cff5")
	err = key.SetId(id, 3)
	c.Assert(err, Equals, ErrInvalidDnetID)
	err = key.SetRawId(id)
	c.Assert(err, Equals, ErrInvalidDnetID)

	id = id[:64]
	err = key.SetId(id, 3)
	c.Assert(err, IsNil)
	c.Assert(key.CmpID(id), Equals, 0)

	err = key.SetRawId(id)
	c.Assert(err, IsNil)
	c.Assert(key.CmpID(id), Equals, 0)
}

func (s *KeySuite) TestKeyNewAndFree(c *C) {
	_, err := NewKey(s.badKey)
	c.Assert(err, Equals, InvalidKeyArgument)

	key, err := NewKey(s.goodKey)
	c.Assert(err, IsNil)
	key.Free()
}

func (s *KeySuite) TestKeyFromIDAndFree(c *C) {
	const id = "21b4f4bd9e64ed355c3eb676a28ebedaf6d8f17bdc365995b319097153044080516bd083bfcce66121a3072646994c8430cc382b8dc543e84880183bf856cff5"
	key, err := NewKeyFromIdStr(id)
	c.Assert(err, IsNil)
	defer key.Free()
	c.Assert(key.ById(), Equals, true)
}

func (s *KeySuite) TestKeysNewAndFree(c *C) {
	keys, err := NewKeys([]string{"A", "B", "C"})
	c.Assert(err, IsNil)
	defer keys.Free()
}

func (s *KeySuite) TestDnetRawIDKeysNewAndFree(c *C) {
	ids := []DnetRawID{
		DnetRawID{[]byte("21b4f4bd9e64ed355c3eb676a28ebedaf6d8f17bdc365995b319097153044080516bd083bfcce66121a3072646994c8430cc382b8dc543e84880183bf856cff5")},
		DnetRawID{[]byte("848b0779ff415f0af4ea14df9dd1d3c29ac41d836c7808896c4eba19c51ac40a439caf5e61ec88c307c7d619195229412eaa73fb2a5ea20d23cc86a9d8f86a0f")},
	}
	keys, err := NewDnetRawIDKeys(ids)
	c.Assert(err, IsNil)
	defer keys.Free()
	c.Assert(keys.Size(), Equals, len(ids))

	keys.InsertID(&DnetRawID{[]byte("3d637ae63d59522dd3cb1b81c1ad67e56d46185b0971e0bc7dd2d8ad3b26090acb634c252fc6a63b3766934314ea1a6e59fa0c8c2bc027a7b6a460b291cd4dfb")})
	c.Assert(keys.Size(), Equals, len(ids)+1)
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
