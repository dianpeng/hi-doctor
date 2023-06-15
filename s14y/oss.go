package s14y

import (
	"github.com/dianpeng/hi-doctor/oss"
	"github.com/dianpeng/hi-doctor/run"

	"crypto/md5"
	"encoding/hex"

	"fmt"
	"io"
	"path"
)

type ossSd struct {
	prefix string
	bucket string
	client oss.Oss
}

func (f *ossSd) readObj(path string) (string, error) {
	body, err := f.client.Get(path)
	if err != nil {
		return "", err
	}
	reader := body.Reader()
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *ossSd) md5(data string) string {
	hash := md5.Sum([]byte(data))
	val := hex.EncodeToString(hash[:])
	return val
}

func (f *ossSd) Refresh(currentJ []Job) []NewJob {
	keyList, err := f.client.List(f.prefix, 10000)
	if err != nil {
		return nil
	}

	targetList := make(map[string]NewJob)
	for _, entry := range keyList {
		ext := path.Ext(entry)
		if ext == ".yaml" {
			data, err := f.readObj(entry)
			if err != nil {
				continue
			}
			name, err := run.GetInspectionName(data)
			if err != nil {
				continue
			}
			targetList[name] = NewJob{
				Name:   name,
				Md5:    f.md5(data),
				Data:   data,
				Origin: fmt.Sprintf("oss://%s/%s", f.bucket, entry),
				Delete: false,
			}
		}
	}

	currentLookup := make(map[string]Job)
	for _, v := range currentJ {
		currentLookup[v.Name] = v
	}
	out := []NewJob{}
	for _, entry := range targetList {
		if old, ok := currentLookup[entry.Name]; ok {
			if old.Md5 == entry.Md5 {
				continue
			}
		}
		out = append(out, entry)
	}

	for _, existed := range currentJ {
		if _, has := targetList[existed.Name]; !has {
			out = append(out, NewJob{
				Name:   existed.Name,
				Delete: true,
			})
		}
	}

	return out
}

func newOss(cfg Config) S14y {
	prefix := cfg.GetStringField("prefix")
	bucket := cfg.GetStringField("bucket")
	provider := cfg.GetStringField("provider")
	factory := oss.GetOSSFactory(provider)
	if factory == nil {
		return nil
	}
	client, err := factory.Create(oss.Option(cfg))
	if err != nil {
		return nil
	}
	return &ossSd{
		prefix: prefix,
		bucket: bucket,
		client: client,
	}
}
