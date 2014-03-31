package ellipticsS3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type riftBackend struct {
	endpoint string
	client   *http.Client
}

// Add different types of errors

func (r *riftBackend) GetObject(key, bucket string) (data []byte, err error) {
	urlStr := fmt.Sprintf("http://%s/get?namespace=%s&name=%s", r.endpoint, bucket, key)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = NewKeyNotFoundError(key, bucket)
		return
	}

	// TODO: add error handler based on StatusCode
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)

	return
}

func (r *riftBackend) UploadObject(key, bucket string, body []byte) (err error) {
	urlStr := fmt.Sprintf("http://%s/upload?namespace=%s&name=%s", r.endpoint, bucket, key)

	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// encode JSON
	_, err = ioutil.ReadAll(resp.Body)
	return
}

func (r *riftBackend) ObjectExists(key, bucket string) (exists bool, err error) {
	urlStr := fmt.Sprintf("http://%s/get?namespace=%s&name=%s&size=1", r.endpoint, bucket, key)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	exists = resp.StatusCode == 200
	return
}

func (r *riftBackend) CreateBucket(bucket string) (err error) {
	task := struct {
		Key       string `json:"key"`
		Groups    []int  `json:"groups"`
		Flags     int    `json:"flags"`
		MaxSize   int    `json:"max-size"`
		MaxKeyNum int    `json:"max-key-num"`
	}{
		Key:       bucket,
		Groups:    []int{1}, // Make it configurable
		Flags:     0,
		MaxSize:   0,
		MaxKeyNum: 0,
	}

	body, _ := json.Marshal(task)

	urlStr := fmt.Sprintf("http://%s/bucket-meta-create", r.endpoint)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Unable to create bucket %s", bucket)
	}

	return
}

func NewRiftbackend(endpoint string) (s S3Backend, err error) {
	// check endpoint by pinging Rift proxy
	pingUrl := fmt.Sprintf("http://%s/ping", endpoint)
	resp, err := http.Get(pingUrl)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Rift is unavailable %s", endpoint)
	}

	//
	s = &riftBackend{
		endpoint: endpoint,
		client:   &http.Client{},
	}
	return
}
