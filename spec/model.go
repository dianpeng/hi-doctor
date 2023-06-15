package spec

import (
	"github.com/dianpeng/hi-doctor/fetch"
	"time"
)

type Storage map[string]string
type Local map[string]string
type Global map[string]string

type Target struct {
	Fetch  *fetch.Fetch             `yaml:"fetch"`
	Format string                   `yaml:"format"`
	Count  uint                     `yaml:"count"`
	Inline []map[string]interface{} `yaml:"inline"`
}

type Task []*TaskAny

type TaskOption map[string]interface{}

type TaskAny struct {
	Guard  string     `yaml:"string"`
	Type   string     `yaml:"type"`
	Option TaskOption `yaml:"option"`
	Check  *Check     `yaml:"check"`
}

type Check struct {
	Condition string   `yaml:"condition"`
	Then      []string `yaml:"then"`
	Otherwise []string `yaml:"otherwise"`
	Lastly    []string `yaml:"lastly"`
}

// External info not derived from yaml but from the runtime
type Info struct {
	Timestamp time.Time
	Md5       string
	Origin    string // if this yaml is been loaded from file, then it is its
	Source    string // source code
}

type MetricItem struct {
	Name  string                 `yaml:"name"`
	Key   string                 `yaml:"key"`
	Type  string                 `yaml:"type"`
	Label map[string]interface{} `yaml:"label"`
}

type Metrics struct {
	Provider  string                 `yaml:"provider"`
	Namespace string                 `yaml:"namespace"`
	Option    map[string]interface{} `yaml:"option"`
	Define    []MetricItem           `yaml:"define"`
}

// Model of the *inspection* job. The model is been assumed to be derived from
// the yaml parser. User are expeceted to use yaml to define and then execute
// by our test runtime
type Model struct {
	Name      string   `yaml:"name"`
	Guard     string   `yaml:"guard"`
	Comment   string   `yaml:"comment"`
	LogPrefix string   `yaml:"log_prefix"`
	Metrics   *Metrics `ymal:"metrics"`
	Storage   Storage  `yaml:"storage"`
	Global    Global   `yaml:"global"`
	Local     Local    `yaml:"local"`
	Trigger   string   `yaml:"trigger"`
	Target    *Target  `yaml:"target"`
	Scheduler string   `yaml:"scheduler"`
	Task      Task     `yaml:"task"`
	Finally   []string `yaml:"finally"`
	Info      Info
}
