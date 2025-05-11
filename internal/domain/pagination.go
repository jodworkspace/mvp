package domain

type Pagination struct {
	Page         uint64            `json:"page"`
	PageSize     uint64            `json:"page_size"`
	Filter       map[string]string `json:"filter"`
	IncludeTotal bool              `json:"include_total"`
}
