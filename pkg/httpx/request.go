package httpx

import "net/url"

func BuildURL(base string, query ...map[string]string) (string, error) {
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
