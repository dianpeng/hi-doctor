package fetch

import (
	"fmt"
	"net/url"

	"github.com/dianpeng/hi-doctor/dvar"
)

type Fetch struct {
	Uri    string                 `yaml:"uri"`
	Option map[string]interface{} `yaml:"option"` // opaque option
}

// An interface to perform the actual *fetch* operation
type Fetcher interface {

	// obtain the resources, returns a byte array which contains all the needed
	// data or an error
	Obtain() ([]byte, error)

	Description() string
}

type FetcherFactory interface {
	Create(*dvar.EvalEnv) (Fetcher, error)
}

type FetcherFactoryCompiler interface {
	Compile(*url.URL, *Fetch) (Fetcher, error)
}

type fetchFactory struct {
	uri   dvar.DVar
	model *Fetch
}

func Compile(f *Fetch) (FetcherFactory, error) {
	uri := f.Uri
	out := &fetchFactory{}

	// fetch.URI
	if dv, err := dvar.NewDVarStringContext(uri); err != nil {
		return nil, err
	} else {
		out.uri = dv
	}

	out.model = f
	return out, nil
}

func (f *fetchFactory) Create(env *dvar.EvalEnv) (Fetcher, error) {
	output, err := f.uri.Value(env)
	if err != nil {
		return nil, fmt.Errorf("fetch_factory.Create fail %s", err)
	}

	urlStr := output.String()

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("fetch_factory.Create url is invalid: %s", err)
	}

	ffc, ok := fetcherFactoryCompiler[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("fetch_factory.Create url scheme is unknown to us")
	}

	return ffc.Compile(u, f.model)
}

var fetcherFactoryCompiler map[string]FetcherFactoryCompiler = make(map[string]FetcherFactoryCompiler)

func RegisterFetchCompiler(scheme string, ffc FetcherFactoryCompiler) {
	fetcherFactoryCompiler[scheme] = ffc
}
