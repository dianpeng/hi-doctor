package fetch

import (
	"fmt"
	"github.com/dianpeng/hi-doctor/oss"
	"io"
	"net/url"
)

type ossFetcher struct {
	cli   oss.Oss
	path  string
	model *Fetch
}

func (f *ossFetcher) Obtain() ([]byte, error) {
	data, err := f.cli.Get(f.path)
	if err != nil {
		return nil, fmt.Errorf("oss_fetcher cannot load %s, %s", f.path, err)
	}
	body, err := io.ReadAll(data.Reader())
	if err != nil {
		return nil, fmt.Errorf("oss_fetcher cannot load %s, %s", f.path, err)
	}

	return body, nil
}

func (f *ossFetcher) Description() string {
	return fmt.Sprintf("oss_fetcher(%s)", f.path)
}

type ossfetcherfactory struct{}

func (_ *ossfetcherfactory) Compile(n *url.URL, f *Fetch) (Fetcher, error) {
	// The scheme is as following, oss://[bucket-name][/]path. Here the bucket
	// name is been encoded as hostname of the URI/URL for now. This is duplicated
	// from the option, field. We do a sanity check here

	path := n.Path
	bucket := n.Hostname()

	// fetched via Fetch.Option field
	opt := f.Option
	if opt == nil {
		return nil, fmt.Errorf("oss_fetch, option is not specified")
	}
	ossOption := oss.Option(opt)
	if bval, ok := ossOption.GetBucket(); !ok {
		ossOption["bucket"] = bucket
	} else {
		if bval != bucket {
			return nil, fmt.Errorf("oss_fetch sanity check fail, " +
				"the bucket name specified in the option is " +
				"different from bucket name in URL's hostname")
		}
	}

	if provider, ok := ossOption.GetProvider(); !ok {
		return nil, fmt.Errorf("oss_fetcher option does not have provider")
	} else {
		ossF := oss.GetOSSFactory(provider)
		if ossF == nil {
			return nil, fmt.Errorf("oss_fetcher provider %s is unknown", provider)
		}
		if cli, err := ossF.Create(ossOption); err != nil {
			return nil, fmt.Errorf(
				"oss_fetcher provider %s creation failed, %s",
				provider,
				err,
			)
		} else {
			return &ossFetcher{
				cli:   cli,
				path:  path,
				model: f,
			}, nil
		}
	}
}

func init() {
	RegisterFetchCompiler("oss", &ossfetcherfactory{})
}
