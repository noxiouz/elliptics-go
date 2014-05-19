package rift

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RiftClient struct {
	endpoint string
	client   *http.Client
}

func (r *RiftClient) CreateBucketDir(bucket string, options BucketDirectoryOptions) (info Info, err error) {
	urlStr := fmt.Sprintf("http://%s/update-bucket-directory/%s", r.endpoint, bucket)

	body, err := json.Marshal(options)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&info)
	return
}

func (r *RiftClient) ListBucketDirectory(bucket string) (info ListingInfo, err error) {
	urlStr := fmt.Sprintf("http://%s/list-bucket-directory/%s", r.endpoint, bucket)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&info)
	return
}

func (r *RiftClient) DeleteBucketDirectory(bucket string) (err error) {
	urlStr := fmt.Sprintf("http://%s/delete-bucket-directory/%s/", r.endpoint, bucket)

	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return
	}

	return
}

func (r *RiftClient) CreateBucket(bucket string, bucketDir string, options BucketOptions) (info Info, err error) {
	urlStr := fmt.Sprintf("http://%s/update-bucket/%s/%s", r.endpoint, bucketDir, bucket)

	body, err := json.Marshal(options)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&info)
	return
}

func (r *RiftClient) ListBucket(bucket string) (info ListingInfo, err error) {
	urlStr := fmt.Sprintf("http://%s/list/%s/", r.endpoint, bucket)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&info)
	return
}

func (r *RiftClient) DeleteBucket(bucket string) (err error) {
	urlStr := fmt.Sprintf("http://%s/delete-bucket/%s", r.endpoint, bucket)

	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return
	}
	return
}

func (r *RiftClient) GetObject(bucket string, key string, size int64, offset int64) (blob []byte, err error) {
	urlStr := fmt.Sprintf("http://%s/get/%s/%s", r.endpoint, bucket, key)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return
	}

	if size > 0 {
		req.URL.Query().Add("size", fmt.Sprintf("%d", size))
	}
	if offset > 0 {
		req.URL.Query().Add("offset", fmt.Sprintf("%d", offset))
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return
	}
	defer resp.Body.Close()
	blob, err = ioutil.ReadAll(resp.Body)
	return
}

func (r *RiftClient) UploadObject(bucket string, key string, blob []byte) (info Info, err error) {
	urlStr := fmt.Sprintf("http://%s/upload/%s/%s", r.endpoint, bucket, key)

	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(blob))
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&info)
	return
}

func (r *RiftClient) DeleteObject(bucket string, key string) (err error) {
	urlStr := fmt.Sprintf("http://%s/delete/%s/%s", r.endpoint, bucket, key)

	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf(resp.Status)
		return
	}
	return
}

func NewRiftClient(endpoint string) (r *RiftClient, err error) {
	// check endpoint by pinging Rift proxy
	pingUrl := fmt.Sprintf("http://%s/ping/", endpoint)
	resp, err := http.Get(pingUrl)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Rift is unavailable %s", endpoint)
	}

	//
	r = &RiftClient{
		endpoint: endpoint,
		client:   &http.Client{},
	}
	return
}
