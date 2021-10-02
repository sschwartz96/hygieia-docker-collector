package models

type Image struct {
	ID          string        `json:"Id"`
	ParentID    string        `json:"ParentId"`
	RepoTags    []interface{} `json:"RepoTags"`
	RepoDigests []interface{} `json:"RepoDigests"`
	Created     int           `json:"Created"`
	Size        int           `json:"Size"`
	VirtualSize int           `json:"VirtualSize"`
	SharedSize  int           `json:"SharedSize"`
	Labels      struct{}      `json:"Labels"`
	Containers  int           `json:"Containers"`
}
