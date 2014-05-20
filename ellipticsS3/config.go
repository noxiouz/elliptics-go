package ellipticsS3

type Config struct {
	Endpoint       string `json:"endpoint"`
	MetaDataGroups []int  `json:"metadata-groups"`
	DataGroups     []int  `json:"groups"`
}
