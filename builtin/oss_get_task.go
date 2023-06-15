package builtin

import (
	"github.com/dianpeng/hi-doctor/check"
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/oss"
	"github.com/dianpeng/hi-doctor/spec"
	"github.com/dianpeng/hi-doctor/task"
	"github.com/dianpeng/hi-doctor/util"

	"fmt"
	"io"
	"time"
)

// OSS task

type ossGetTemplate struct {
	option oss.Option // opaque
	client oss.Oss    // oss client
	path   dvar.DVar  // path of the object
	check  check.Check
}

type ossGetDefine struct {
	RespErr      string `json:"resp_err"`
	RespOK       bool   `json:"resp_ok"`
	RespObj      string `json:"resp_obj"`
	RespObjSize  int64  `json:"resp_obj_size"`
	RespTruncate bool   `json:"resp_truncate"`
	Timestamp    int64  `json:"timestamp"`
	RespTTFB     int64  `json:"resp_ttfb"`
	RespRT       int64  `json:"resp_rt"`
}

type ossGetTask struct {
	t    *ossGetTemplate
	path string
}

type ossGetTaskFactory struct{}

func newOssGetTask(t *ossGetTemplate) *ossGetTask {
	return &ossGetTask{
		t:    t,
		path: "",
	}
}

// OSSGetTaskFactory

func newOssGetTemplate(
	x spec.TaskOption,
	c *spec.Check,
) (task.TaskPlanner, error) {
	tmpl := &ossGetTemplate{}
	tmpl.option = oss.Option(x)

	if name, ok := tmpl.option.GetProvider(); !ok {
		return nil, fmt.Errorf("oss_get provider field is not in option")
	} else {
		factory := oss.GetOSSFactory(name)
		if factory == nil {
			return nil, fmt.Errorf("oss_get provider(%s) is unknown to us", name)
		}

		if cli, err := factory.Create(tmpl.option); err != nil {
			return nil, fmt.Errorf(
				"oss_get provider %s client creation failed %s",
				name,
				err,
			)
		} else {
			tmpl.client = cli
		}

		if path, ok := tmpl.option.GetPath(); ok {
			if dv, err := dvar.NewDVarStringContext(path); err != nil {
				return nil, fmt.Errorf("oss_get path compiles fail: %s", err)
			} else {
				tmpl.path = dv
			}
		} else {
			tmpl.path = dvar.NewDVarLit("")
		}

		if ck, err := check.CompileCheck(c); err != nil {
			return nil, fmt.Errorf("oss_get check compile fail: %s", err)
		} else {
			tmpl.check = ck
		}
	}

	return tmpl, nil
}

func (f *ossGetTaskFactory) Compile(
	x spec.TaskOption,
	c *spec.Check,
) (task.TaskPlanner, error) {
	return newOssGetTemplate(x, c)
}

func (f *ossGetTaskFactory) SanityCheck(x spec.TaskOption) error {
	opt := oss.Option(x)
	if !opt.HasProvider() {
		return fmt.Errorf("oss_get option does not have provider field")
	}
	return nil
}

// OSSGetTemplate (TaskPlanner)

func (p *ossGetTemplate) Description() string {
	return "oss_get"
}

func (p *ossGetTemplate) GenTask(env *dvar.EvalEnv) (task.TaskList, error) {
	return task.TaskList{
		newOssGetTask(p),
	}, nil
}

// OSSGetTask
func (t *ossGetTask) Prepare(env *dvar.EvalEnv) error {
	if v, err := ossGetPath("oss_get", &t.t.path, env); err != nil {
		return err
	} else {
		t.path = v
	}
	return nil
}

func (t *ossGetTask) doRunGet(env *dvar.EvalEnv) *ossGetDefine {
	stat := &ossGetDefine{}

	startTs := time.Now().UnixMilli()
	obj, err := t.t.client.Get(t.path)
	endTs := time.Now().UnixMilli()

	if err != nil {
		stat.RespErr = fmt.Sprintf("%s", err)
		stat.RespOK = false
		stat.RespObj = ""
		stat.RespObjSize = int64(-1)
		stat.RespTruncate = false
		stat.RespTTFB = endTs - startTs
		stat.RespRT = -1
		stat.Timestamp = startTs

	} else {
		stat.RespErr = ""

		// try to read the whole body out
		body := obj.Reader()
		defer body.Close()

		if data, err := io.ReadAll(body); err != nil {
			stat.RespErr = fmt.Sprintf("%s", err)
			stat.RespOK = false
			stat.RespObj = ""
			stat.RespObjSize = int64(-1)
			stat.RespTruncate = false
			stat.RespTTFB = endTs - startTs
			stat.RespRT = -1
			stat.Timestamp = startTs
		} else {
			endBodyTs := time.Now().UnixMilli()

			stat.RespErr = ""
			stat.RespOK = true
			stat.RespObj = string(data)
			stat.RespObjSize = int64(len(stat.RespObj))
			stat.RespTruncate = false
			stat.RespTTFB = endTs - startTs
			stat.RespRT = endBodyTs - startTs
			stat.Timestamp = startTs
		}
	}

	return stat
}

func (t *ossGetTask) runGet(env *dvar.EvalEnv) error {
	def := t.doRunGet(env)
	stat := util.ToMapInterface(def)
	env.RecordHistoricalResult("oss_get", stat)
	return nil
}

func (t *ossGetTask) Run(env *dvar.EvalEnv) error {
	if err := t.runGet(env); err != nil {
		return err
	}
	if err := t.t.check.Run(env); err != nil {
		return err
	}
	return nil
}

func (t *ossGetTask) Description() string {
	return "oss_get"
}

func init() {
	task.RegisterTaskFactory("oss_get", &ossGetTaskFactory{})
}
