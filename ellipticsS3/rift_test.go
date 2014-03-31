package ellipticsS3

import (
	"testing"
)

const defHost = "cocaine-cloud02g.kit.yandex.net:9000"

func TestRiftGetObject(t *testing.T) {
	rift, err := NewRiftbackend(defHost)
	if err != nil {
		t.Fatal(err)
	}

	data, err := rift.GetObject("xxx.jpg", "testns")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Data size: %d", len(data))

	_, err = rift.GetObject("xxx.jpg", "testns")
	if err == nil {
		t.Fatal(err)
	}

}

func TestConnector(t *testing.T) {
	_, err := NewRiftbackend(defHost)
	if err != nil {
		t.Fatalf("Unexpected error %s", err)
	}

	_, err = NewRiftbackend("dummyHost")
	if err != nil {
		t.Fatalf("Expected error, but it didn't occure %s", err)
	}
}

func TestRiftObjectExists(t *testing.T) {
	rift, err := NewRiftbackend(defHost)
	if err != nil {
		t.Fatal(err)
	}

	exists, err := rift.ObjectExists("xxx.jpg", "testns")
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("Object doesn't exist")
	}

	exists, err = rift.ObjectExists("xxx2.jpg", "testns")
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("Object exists")
	}
}

func TestRiftUploadObject(t *testing.T) {
	rift, err := NewRiftbackend(defHost)
	if err != nil {
		t.Fatal(err)
	}

	err = rift.UploadObject("testObj", "testns", []byte("AAAAAA"))
	if err != nil {
		t.Fatal(err)
	}
}
