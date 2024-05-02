package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type httpFetcher struct {
	baseURL string
}

var _ Fetcher = (*httpFetcher)(nil)

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
	// 从 reader 中读字节切片
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %v", err)
	}

	return bytes, nil
}
