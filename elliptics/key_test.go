package elliptics

import (
	"testing"
)

func TestKey(t *testing.T) {
	t.Log("Start test")
	key, err := NewKey()
	if err != nil {
		t.Errorf("%v", key)
	}

	if key.ById() {
		t.Errorf("%s", "Create key without ID")
	}

}
