package httpx

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Client interface {
	BuildURL(base string, query map[string]string) (string, error)
	DoRequest(ctx context.Context, method, url string, body io.Reader, headers ...http.Header) (*http.Response, error)
}

type client struct {
	http.Client
}

func NewHTTPClient(c http.Client) Client {
	return &client{c}
}

func (h *client) DoRequest(ctx context.Context, method, url string, body io.Reader, headers ...http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
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

func (*client) BuildURL(base string, query map[string]string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	for k, v := range query {
		params.Add(k, v)
	}
	b.RawQuery = params.Encode()

	return b.String(), nil
}
