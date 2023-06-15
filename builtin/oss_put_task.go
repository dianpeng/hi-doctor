package builtin

import (
	"github.com/dianpeng/hi-doctor/check"
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/oss"
	"github.com/dianpeng/hi-doctor/spec"
	"github.com/dianpeng/hi-doctor/task"
	"github.com/dianpeng/hi-doctor/util"

	"fmt"
	"strings"
	"time"
)

// OSS task

type ossPutTemplate struct {
	option oss.Option // opaque
	client oss.Oss    // oss client
	path   dvar.DVar  // path of the object
	object dvar.DVar  // a fake object needs to be upload
	check  check.Check
}

type ossPutDefine struct {
	RespErr   string `json:"resp_err"`
	RespOK    bool   `json:"resp_ok"`
	Timestamp int64  `json:"timestamp"`
	RespRT    int64  `json:"resp_rt"`
}

type ossPutTask struct {
	t      *ossPutTemplate
	path   string
	object string
}

type ossPutTaskFactory struct{}

func newOssPutTask(t *ossPutTemplate) *ossPutTask {
	return &ossPutTask{
		t:      t,
		path:   "",
		object: "",
	}
}

// OSSGetTaskFactory

func newOssPutTemplate(
	x spec.TaskOption,
	c *spec.Check,
) (task.TaskPlanner, error) {
	tmpl := &ossPutTemplate{}
	tmpl.option = oss.Option(x)

	if name, ok := tmpl.option.GetProvider(); !ok {
		return nil, fmt.Errorf("oss_put provider field is not in option")
	} else {
		factory := oss.GetOSSFactory(name)
		if factory == nil {
			return nil, fmt.Errorf("oss_put provider(%s) is unknown to us", name)
		}

		if cli, err := factory.Create(tmpl.option); err != nil {
			return nil, fmt.Errorf(
				"oss_put provider %s client creation failed %s",
				name,
				err,
			)
		} else {
			tmpl.client = cli
		}

		if path, ok := tmpl.option.GetPath(); ok {
			if dv, err := dvar.NewDVarStringContext(path); err != nil {
				return nil, fmt.Errorf("oss_put path compiles fail: %s", err)
			} else {
				tmpl.path = dv
			}
		} else {
			tmpl.path = dvar.NewDVarLit("")
		}

		if obj, ok := tmpl.option.GetObject(); ok {
			if dv, err := dvar.NewDVarStringContext(obj); err != nil {
				return nil, fmt.Errorf("oss_put object compiles fail: %s", err)
			} else {
				tmpl.object = dv
			}
		} else {
			tmpl.object = dvar.NewDVarLit("")
		}

		if ck, err := check.CompileCheck(c); err != nil {
			return nil, fmt.Errorf("oss_put check compile fail: %s", err)
		} else {
			tmpl.check = ck
		}
	}

	return tmpl, nil
}

func (f *ossPutTaskFactory) Compile(
	x spec.TaskOption,
	c *spec.Check,
) (task.TaskPlanner, error) {
	return newOssPutTemplate(x, c)
}

func (f *ossPutTaskFactory) SanityCheck(x spec.TaskOption) error {
	opt := oss.Option(x)
	if !opt.HasProvider() {
		return fmt.Errorf("oss_put option does not have provider field")
	}
	return nil
}

// OSSGetTemplate (TaskPlanner)

func (p *ossPutTemplate) Description() string {
	return "oss_put"
}

func (p *ossPutTemplate) GenTask(env *dvar.EvalEnv) (task.TaskList, error) {
	return task.TaskList{
		newOssPutTask(p),
	}, nil
}

// OSSGetTask
func (t *ossPutTask) Prepare(env *dvar.EvalEnv) error {
	if v, err := ossGetPath("oss_put", &t.t.path, env); err != nil {
		return err
	} else {
		t.path = v
	}

	if v, err := ossGetObject("oss_put", &t.t.object, env); err != nil {
		return err
	} else {
		t.object = v
	}

	return nil
}

func (t *ossPutTask) doRunGet(env *dvar.EvalEnv) *ossPutDefine {
	stat := &ossPutDefine{}

	startTs := time.Now().UnixMilli()
	err := t.t.client.Put(t.path, strings.NewReader(t.object), int64(len(t.object)))
	endTs := time.Now().UnixMilli()

	if err != nil {
		stat.RespErr = fmt.Sprintf("%s", err)
		stat.RespOK = false
		stat.RespRT = -1
		stat.Timestamp = startTs
	} else {
		stat.RespErr = ""
		stat.RespOK = true
		stat.RespRT = endTs - startTs
		stat.Timestamp = startTs
	}

	return stat
}

func (t *ossPutTask) runGet(env *dvar.EvalEnv) error {
	def := t.doRunGet(env)
	stat := util.ToMapInterface(def)
	env.RecordHistoricalResult("oss_put", stat)
	return nil
}

func (t *ossPutTask) Run(env *dvar.EvalEnv) error {
	if err := t.runGet(env); err != nil {
		return err
	}
	if err := t.t.check.Run(env); err != nil {
		return err
	}
	return nil
}

func (t *ossPutTask) Description() string {
	return "oss_put"
}

func init() {
	task.RegisterTaskFactory("oss_put", &ossPutTaskFactory{})
}
