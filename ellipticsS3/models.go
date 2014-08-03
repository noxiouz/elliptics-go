package ellipticsS3

import (
	"encoding/xml"
)

// Prototype structs to generate XML result of listing operations

type Owner struct {
	Id          string `xml:"Owner>ID"`
	DisplayName string `xml:"Owner>DisplayName"`
}

type XMLBucketDirectoryList struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Owner
	Buckets []XMLBucketItem `xml:"Bucket"`
}

type XMLBucketItem struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

type XMLContentItem struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	Owner
}

type XMLBucketList struct {
	XMLName     xml.Name         `xml:"ListBucketResult"`
	Name        string           `xml:"Name"`
	MaxKeys     int              `xml:"MaxKeys"`
	IsTruncated bool             `xml:"IsTruncated"`
	Contents    []XMLContentItem `xml:"Contents"`
}
