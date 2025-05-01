package middleware

type Pagination struct {
	Page         int               `json:"page"`
	PageSize     int               `json:"page_size"`
	Filter       map[string]string `json:"filter"`
	IncludeTotal bool              `json:"include_total"`
}
