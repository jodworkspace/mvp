package middleware

type pagination struct {
	Page         int  `json:"page"`
	PageSize     int  `json:"page_size"`
	IncludeTotal bool `json:"include_total"`
}
