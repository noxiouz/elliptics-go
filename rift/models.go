package rift

type ACLStruct struct {
	User  string `json:"user"`
	Token string `json:"token"`
	Flags int    `json:"flags"`
}

type BucketOptions struct {
	Groups    []int       `json:"groups"`
	ACL       []ACLStruct `json:"acl"`
	Flags     int         `json:"flags"`
	MaxSize   int         `json:"max-size"`
	MaxKeyNum int         `json:"max-key-num"`
}

type BucketDirectoryOptions BucketOptions

//
type FileGroupInfo struct {
	CSum     string `json:"csum"`
	Id       string `json:"id"`
	Filename string `json:"filename"`
	Offset   int64  `json:"offset-within-data-file"`
	MTime    struct {
		Time    string `json:"time"`
		RawTime string `json:"time-raw"`
	} `json:"mtime`
	Server string `json:"server"`
	Size   int    `json:"size"`
}

type Info struct {
	Info []FileGroupInfo `json:"info"`
}

//
type ListingInfo struct {
	Indexes []ItemIndexInfo `json:"indexes"`
}

func (l *ListingInfo) Keys() (keys []string) {
	for _, item := range l.Indexes {
		keys = append(keys, item.Key)
	}
	return
}

type ItemIndexInfo struct {
	Id          string `json:"id"`
	Key         string `json:"key"`
	Timestamp   string `json:"timestamp"`
	TimeSeconds string `json:"time_seconds"`
}
