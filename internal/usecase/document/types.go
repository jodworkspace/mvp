package document

type File struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"mimeType"`
}

type FileListResponse struct {
	IncompleteSearch bool         `json:"incompleteSearch"`
	NextPageToken    string       `json:"nextPageToken"`
	Files            []File       `json:"files"`
	Kind             string       `json:"kind"`
	Error            *GoogleError `json:"error,omitempty"`
}

type GoogleError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Errors  []struct {
		Message      string `json:"message"`
		Domain       string `json:"domain"`
		Reason       string `json:"reason"`
		Location     string `json:"location"`
		LocationType string `json:"locationType"`
	} `json:"errors"`
	Status string `json:"status"`
}
