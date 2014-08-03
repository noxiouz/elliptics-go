package ellipticsS3

import "fmt"

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

type Context struct {
	Username string
}
