package domain

type Document struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsFolder bool   `json:"isFolder"`
}
