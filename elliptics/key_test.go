package elliptics

import (
	"crypto/sha512"
	"errors"
	"testing"
)

/*
	Key
*/

const (
	badKeyCreationArg  = 9999
	goodKeyCreationArg = "some_key"
)

func TestKeyDefaultCreationAndFree(t *testing.T) {
	key, err := NewKey()
	if err != nil {
		t.Errorf("%v", key)
	}

	// t.Log(key.CmpID([]uint8{1, 2, 3, 4, 5}))

	if key.ById() {
		t.Errorf("%s", "Create key without ID")
	}

	key.Free()
}

func TestKeyCreationAndFree(t *testing.T) {
	_, err := NewKey(badKeyCreationArg)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	key, err := NewKey(goodKeyCreationArg)
	if err != nil {
		t.Fatalf("Error in a key creation, got %v", err)
	}
	key.Free()
}

/*
	Keys
*/

func TestKeysCreationAndFree(t *testing.T) {
	t.Skip("Skip this test")
	keys, err := NewKeys([]string{"A", "B", "C"})
	if err != nil {
		t.Fatalf("NewKeys: Unexpected error %s", err)
	}
	defer keys.Free()

	var hash []uint8
	for _, v := range sha512.Sum512([]byte("A")) {
		hash = append(hash, v)
	}
	name, err := keys.Find(hash)
	if err != nil {
		t.Errorf("Find: Unexpected error %s", err)
	}

	if name != "A" {
		t.Errorf("Unexpected `name` value %s", name)
	}
}

/*
	Error
*/

func TestDnetError(t *testing.T) {
	const (
		dnetCode = 100
		dnetFlag = 16
		dnetMsg  = "dummy_dnet_error_message"
	)
	derr := DnetError{
		Code:    dnetCode,
		Flags:   dnetFlag,
		Message: dnetMsg,
	}

	dummyErr := errors.New("dummy_err")

	if msg := ErrorData(&derr); msg != dnetMsg {
		t.Errorf("ErroData: expected %s, got %s", dnetMsg, msg)
	}

	if msg := ErrorData(dummyErr); msg != dummyErr.Error() {
		t.Errorf("ErroData: expected %s, got %s", dummyErr.Error(), msg)
	}

	if len(derr.Error()) == 0 {
		t.Errorf("DnetError: a malformed error representation")
	}

}
