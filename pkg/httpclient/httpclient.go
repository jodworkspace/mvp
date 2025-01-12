package httpclient

import "net/http"

type HTTPClient struct {
	*http.Client
}
