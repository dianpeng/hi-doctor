package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type httpFetcher struct {
	url *url.URL // url object

	// for now we do not support any sort of complicated method, in the future
	// we can somehow add some simple authentication, like Basic, Digest, even
	// JWT ect ...
	model *Fetch
}

func (h *httpFetcher) Obtain() ([]byte, error) {
	resp, err := http.Get(h.url.String())
	if err != nil {
		return nil, fmt.Errorf("http fetcher(%s) fail %s", h.url.String(), err)
	}

	// read the full body out
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (h *httpFetcher) Description() string {
	return fmt.Sprintf("http_fetch(%s)", h.url.String())
}

func compileHttpFetcher(url *url.URL, model *Fetch) (Fetcher, error) {
	return &httpFetcher{
		url:   url,
		model: model,
	}, nil
}

type httpfetcherfactory struct {
}

func (_ *httpfetcherfactory) Compile(n *url.URL, f *Fetch) (Fetcher, error) {
	return compileHttpFetcher(n, f)
}

func init() {
	RegisterFetchCompiler("http", &httpfetcherfactory{})
	RegisterFetchCompiler("https", &httpfetcherfactory{})
}
