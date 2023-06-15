package fetch

import (
	"fmt"
	"io"
	"net/url"
	"os"
)

type fileFetcher struct {
	path  string
	model *Fetch
}

func (f *fileFetcher) Obtain() ([]byte, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, fmt.Errorf("file_fetcher(%s) cannot open file: %s", f.path, err)
	}
	defer file.Close()

	return io.ReadAll(file)
}

func (f *fileFetcher) Description() string {
	return fmt.Sprintf("file_fetcher(%s)", f.path)
}

func compileFileFetcher(url *url.URL, model *Fetch) (Fetcher, error) {
	path := url.String()[7:]
	return &fileFetcher{
		path:  path,
		model: model,
	}, nil
}

type filefetcherfactory struct{}

func (_ *filefetcherfactory) Compile(n *url.URL, f *Fetch) (Fetcher, error) {
	return compileFileFetcher(n, f)
}

func init() {
	RegisterFetchCompiler("file", &filefetcherfactory{})
}
