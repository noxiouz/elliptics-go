package ellipticsS3

type ObjectController interface {
	GetObject(key, bucket string) ([]byte, error)
	UploadObject(key, bucket string, body []byte) error
	ObjectExists(key, bucket string) (exists bool, err error)
}

type S3Backend interface {
	ObjectController
}
