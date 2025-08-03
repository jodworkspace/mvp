package domain

type Filter struct {
	Page       uint64         `json:"page"`
	PageSize   uint64         `json:"pageSize"`
	Conditions map[string]any `json:"conditions"`
}
