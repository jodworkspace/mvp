package domain

type Pagination struct {
	Page      uint64 `json:"page"`
	PageToken string `json:"pageToken"`
	PageSize  uint64 `json:"pageSize"`
}
