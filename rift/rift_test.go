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
		Groups:    []int{1},
		ACL:       make([]ACLStruct, 0),
		Flags:     0,
		MaxSize:   0,
		MaxKeyNum: 0,
	}
	testBucketDirOpt BucketDirectoryOptions = BucketDirectoryOptions{
		Groups:    []int{1},
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

	if keys := lst.Keys(); len(keys) == 0 || testBucket != keys[0] {
		t.Fatalf("TestBucket is not found in listing %s %v", testBucket, lst)
	}

	err = r.DeleteBucketDirectory(testDir)
	if err != nil {
		t.Fatal(err)
	}
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

	// listing, err := r.ListBucket(testBucket)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// t.Logf("Listing %s", listing)

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
		t.Fatal(err)
	}

	info, err := r.UploadObject(testBucket, testKey, blob)
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Info) == 0 {
		t.Fatalf("Wrong info %# v", pretty.Formatter(info))
	}

	value, err := r.GetObject(testBucket, testKey)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(value, blob) {
		t.Fatal("value is not equal to blob %s %s", value, blob)
	}

	err = r.DeleteObject(testBucket, testKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = r.GetObject(testBucket, testKey)
	if err == nil {
		t.Fatalf("Error is expected %s", err)
	}
}
