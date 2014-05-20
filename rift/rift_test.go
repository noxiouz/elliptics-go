package rift

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kr/pretty"
)

var (
	_ = fmt.Sprintf
	_ = time.Now
)

var (
	riftEndpoint  string
	testDir       string        = fmt.Sprintf("testdir-%d", time.Now().Unix())
	testBucket    string        = fmt.Sprintf("testbucket-%d", time.Now().Unix())
	testKey       string        = fmt.Sprintf("testkey-%d", time.Now().Unix())
	testBucketOpt BucketOptions = BucketOptions{
		Groups:    []int{2},
		ACL:       make([]ACLStruct, 0),
		Flags:     0,
		MaxSize:   0,
		MaxKeyNum: 0,
	}
	testBucketDirOpt BucketDirectoryOptions = BucketDirectoryOptions{
		Groups:    []int{2},
		ACL:       make([]ACLStruct, 0),
		Flags:     0,
		MaxSize:   0,
		MaxKeyNum: 0,
	}
)

func init() {
	riftEndpoint = os.Getenv("RIFT_ENDPOINT")
}

func TestEnv(t *testing.T) {
	if riftEndpoint == "" {
		t.Fatal("Set environment variable RIFT_ENDPOINT")
	}
}

func TestBucketDir(t *testing.T) {
	if _, err := NewRiftClient("SOMERANDOMENDPOINT:10000"); err == nil {
		t.Fatal("Error expected, as unreal endpoint has been used")
	}

	r, err := NewRiftClient(riftEndpoint)
	if err != nil {
		t.Fatal(err)
	}

	info, err := r.CreateBucketDir(testDir, testBucketDirOpt)
	if err != nil {
		t.Fatal(err)
	}

	if len(info.Info) == 0 {
		t.Fatalf("Wrong info %# v", pretty.Formatter(info))
	}

	_, err = r.CreateBucket(testBucket, testDir, testBucketOpt)
	if err != nil {
		t.Fatal(err)
	}

	lst, err := r.ListBucketDirectory(testDir)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(lst.Keys())

	if keys := lst.Keys(); len(keys) == 0 || testBucket != keys[0] {
		t.Fatalf("TestBucket is not found in listing %s %v", testBucket, lst)
	}

	// err = r.DeleteBucketDirectory(testDir)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}

func TestBucket(t *testing.T) {
	r, err := NewRiftClient(riftEndpoint)
	if err != nil {
		t.Fatal(err)
	}

	r.CreateBucketDir(testDir, testBucketDirOpt)
	if err != nil {
		t.Fatal("Create error", err)
	}

	info, err := r.CreateBucket(testBucket, testDir, testBucketOpt)
	if err != nil {
		t.Fatal(err)
	}

	if len(info.Info) == 0 {
		t.Fatalf("Wrong info %# v", pretty.Formatter(info))
	}

	bucketInfo, err := r.ReadBucket(testBucket)
	if err != nil {
		t.Fatalf("Unable to read bucket %s", testBucket)
	}

	t.Log(bucketInfo)

	_, err = r.ListBucket(testBucket)
	if err != nil {
		t.Fatalf("Unable to list directory %s %s", testBucket, err)
	}

	// err = r.DeleteBucket(testBucket)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}

func TestObject(t *testing.T) {
	r, err := NewRiftClient(riftEndpoint)
	if err != nil {
		t.Fatal(err)
	}

	blob := []byte("TESTBLOB")
	_, err = r.CreateBucket(testBucket, testDir, testBucketOpt)
	if err != nil {
		t.Fatalf("Unable to create bucket %s", err)
	}

	info, err := r.UploadObject(testBucket, testKey, blob)
	if err != nil {
		t.Fatalf("Unable to upload object %s", err)
	}
	if len(info.Info) == 0 {
		t.Fatalf("Wrong info %# v", pretty.Formatter(info))
	}

	listing, err := r.ListBucket(testBucket)
	if err != nil {
		t.Fatalf("Unable to list directory %s %s", testBucket, err)
	}
	t.Log(listing.Keys())

	value, err := r.GetObject(testBucket, testKey, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(value, blob) {
		t.Fatal("value is not equal to blob %s %s", value, blob)
	}

	secondByte, err := r.GetObject(testBucket, testKey, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(secondByte) == 0 || secondByte[0] != blob[1] {
		t.Fatalf("%s %s", secondByte[0], blob[1])
	}

	err = r.DeleteObject(testBucket, testKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = r.GetObject(testBucket, testKey, 0, 0)
	if err == nil {
		t.Fatalf("Error is expected %s", err)
	}
}

func TestExtermination(t *testing.T) {
	r, err := NewRiftClient(riftEndpoint)
	if err != nil {
		t.Fatal(err)
	}

	err = r.DeleteBucket(testBucket)
	if err != nil {
		t.Fatalf("Unable to delete directory %s %s", testBucket, err)
	}

	lst, err := r.ListBucketDirectory(testDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(lst.Keys()) != 0 {
		t.Fatal("Bucket dir %s must be empty, but %s", testDir, lst.Keys())
	}

	err = r.DeleteBucketDirectory(testDir)
	if err != nil {
		t.Fatal("Unable to delete bucket directory %s %s", testDir, err)
	}

	if _, err := r.ListBucketDirectory(testDir); err == nil {
		t.Fatal("Expected error, but got nil %s", testDir)
	}

}
