package httpclient

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type HTTPClient struct {
	*http.Client
}

func NewHTTPClient(c *http.Client) *HTTPClient {
	return &HTTPClient{c}
}

func (h *HTTPClient) SendJSONRequest(method, url string, payload any, headers ...http.Header) (*http.Response, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if len(headers) > 0 {
		for k, v := range headers[0] {
			req.Header[k] = v
		}
	}

	return h.Do(req)
}

func (h *HTTPClient) SendGetRequest(url string, headers ...http.Header) (*http.Response, error) {
	return h.SendJSONRequest(http.MethodGet, url, nil, headers...)
}
