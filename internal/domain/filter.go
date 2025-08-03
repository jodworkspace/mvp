package domain

type Filter struct {
	Page       uint64         `json:"page"`
	PageSize   uint64         `json:"page_size"`
	Conditions map[string]any `json:"conditions"`
}
