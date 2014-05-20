package ellipticsS3

import (
	"encoding/xml"
)

type XMLBucketDirectoryList struct {
	XMLName     xml.Name        `xml:"ListAllMyBucketsResult"`
	Id          string          `xml:"Owner>ID"`
	DisplayName string          `xml:"Owner>DisplayName"`
	Buckets     []XMLBucketItem `xml:"Bucket"`
}

type XMLBucketItem struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}
