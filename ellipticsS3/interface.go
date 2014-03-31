package ellipticsS3

import "fmt"

type ObjectController interface {
	GetObject(key, bucket string) ([]byte, error)
	UploadObject(key, bucket string, body []byte) error
	ObjectExists(key, bucket string) (exists bool, err error)
}

type BucketController interface {
	CreateBucket(bucket string) error
}

type S3Backend interface {
	ObjectController
	BucketController
}

// Errors

type KeyNotFoundError struct {
	key    string
	bucket string
}

func (k *KeyNotFoundError) Error() string {
	return fmt.Sprintf("Key '%s' wasn't discovered in bucket '%s'", k.key, k.bucket)
}

func NewKeyNotFoundError(key, bucket string) error {
	return &KeyNotFoundError{
		key:    key,
		bucket: bucket,
	}
}
