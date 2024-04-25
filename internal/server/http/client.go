package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"recache/internal/service"
)

type httpFetcher struct {
	baseURL string
}

var _ service.Fetcher = (*httpFetcher)(nil)

func (h *httpFetcher) Fetch(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	fmt.Println("查询的key：", string(key))
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %v", err)
	}

	return bytes, nil
}
