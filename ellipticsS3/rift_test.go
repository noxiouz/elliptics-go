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

	_, err = rift.GetObject("xxaa.jpg", "testns")
	if err == nil {
		t.Fatal(err)
	} else {
		t.Log(err)
	}

}

func TestConnector(t *testing.T) {
	_, err := NewRiftbackend(defHost)
	if err != nil {
		t.Fatalf("Unexpected error %s", err)
	}

	_, err = NewRiftbackend("dummyHost")
	if err == nil {
		t.Fatalf("Expected error, but it didn't occure %s", err)
	}
}

func TestRiftObjectExists(t *testing.T) {
	rift, err := NewRiftbackend(defHost)
	if err != nil {
		t.Fatal(err)
	}

	exists, err := rift.ObjectExists("random_key", "random_bucket")
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

	objName := "testObj"
	bucketName := "testns"
	testData := []byte("dummy_test_data")
	err = rift.UploadObject(objName, bucketName, testData)
	if err != nil {
		t.Fatal(err)
	}

	exists, err := rift.ObjectExists(objName, bucketName)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("Object doesn't exist")
	}

	data, err := rift.GetObject(objName, bucketName)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) != len(testData) {
		t.Fatalf("Get error. Expected %s ,but %s got", testData, data)
	}
}
