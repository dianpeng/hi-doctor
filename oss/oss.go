package oss

import (
	"io"
	"net/http"
	"time"
)

type Option map[string]interface{}

type Object interface {
	Reader() io.ReadCloser
	Size() int64
	ModifyTime() time.Time
	Headers() http.Header
}

type Oss interface {
	Get(string) (Object, error)
	Put(string, io.Reader, int64) error
	Del(string) error
	Provider() string
	// list all the files in the current directory
	List(string, int) ([]string, error)
}

type OssFactory interface {
	Create(Option) (Oss, error)
}

func (o *Option) getString(x string) (string, bool) {
	if v, ok := (*o)[x]; ok {
		if vv, ok := v.(string); ok {
			return vv, true
		}
	}
	return "", false
}

func (o *Option) HasPath() bool {
	_, v := o.GetPath()
	return v
}

func (o *Option) GetPath() (string, bool) {
	return o.getString("path")
}

func (o *Option) HasProvider() bool {
	_, v := o.GetProvider()
	return v
}

func (o *Option) GetProvider() (string, bool) {
	return o.getString("provider")
}

func (o *Option) HasObject() bool {
	_, v := o.GetObject()
	return v
}

func (o *Option) GetObject() (string, bool) {
	return o.getString("object")
}

func (o *Option) HasBucket() bool {
	_, ok := o.GetBucket()
	return ok
}

func (o *Option) GetBucket() (string, bool) {
	return o.getString("bucket")
}

func (o *Option) HasAccessKey() bool {
	_, ok := o.GetAccessKey()
	return ok
}

func (o *Option) GetAccessKey() (string, bool) {
	return o.getString("access_key")
}

var (
	facMap = make(map[string]OssFactory)
)

func RegisterOSSFactory(name string, f OssFactory) {
	facMap[name] = f
}

func GetOSSFactory(name string) OssFactory {
	v, _ := facMap[name]
	return v
}
