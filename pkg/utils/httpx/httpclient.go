package httpx

import (
	"net/http"
	"net/url"
)

type Client interface {
	BuildURL(base string, query ...map[string]string) (string, error)
	DoRequest(method, url string, headers ...http.Header) (*http.Response, error)
	DoRequestWithJSONBody(method, url string, body any, headers ...http.Header) (*http.Response, error)
}

type httpClient struct {
	*http.Client
}

func NewHTTPClient(c *http.Client) Client {
	return &httpClient{c}
}

func (h *httpClient) DoRequest(method, url string, headers ...http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			req.Header[k] = v
		}
	}

	return h.Do(req)
}

func (h *httpClient) DoRequestWithJSONBody(method, url string, body any, headers ...http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
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

func (*httpClient) BuildURL(base string, query ...map[string]string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	if len(query) > 0 {
		params := url.Values{}
		for k, v := range query[0] {
			params.Add(k, v)
		}
		b.RawQuery = params.Encode()
	}

	return b.String(), nil
}
