package builtin

import (
	"github.com/dianpeng/hi-doctor/check"
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/spec"
	"github.com/dianpeng/hi-doctor/task"
	"github.com/dianpeng/hi-doctor/util"

	"github.com/mitchellh/mapstructure"

	"fmt"
	"net"
	"time"
)

type tcpTaskTemplate struct {
	name    string
	timeout int64
	check   check.Check
}

type tcpTask struct {
	t         *tcpTaskTemplate
	address   string
	portRange []uint16
}

type tcpTaskDefine struct {
	Name    string `mapstructure:"name"`
	Timeout int64  `mapstructure:"timeout"`
}

type tcpTaskResult struct {
	Timestamp    int64  `json:"timestamp"`
	RT           int64  `json:"rt"`
	Port         uint16 `json:"port"`
	Address      string `json:"address"`
	LocalAddress string `json:"local_address"`
	OK           bool   `json:"ok"`
	Error        string `json:"error"`
}

type tcpTaskFactory struct{}

func (f *tcpTaskFactory) SanityCheck(spec.TaskOption) error {
	return nil
}

func (f *tcpTaskFactory) Compile(
	x spec.TaskOption,
	c *spec.Check,
) (task.TaskPlanner, error) {
	opt := &tcpTaskDefine{
		Timeout: 30,
	}
	err := mapstructure.Decode(opt, opt)
	if err != nil {
		return nil, fmt.Errorf("tcp_task, invalid option input: %s", err)
	}

	out := &tcpTaskTemplate{
		name:    opt.Name,
		timeout: opt.Timeout,
	}

	if ck, err := check.CompileCheck(c); err != nil {
		return nil, fmt.Errorf("tcp_task, check compilation fail: %s", err)
	} else {
		out.check = ck
	}
	return out, nil
}

func (f *tcpTaskTemplate) Description() string {
	return fmt.Sprintf("tcp_task(%s)", f.name)
}

func (f *tcpTaskTemplate) GenTask(env *dvar.EvalEnv) (task.TaskList, error) {
	return task.TaskList{
		&tcpTask{
			t: f,
		},
	}, nil
}

func (t *tcpTask) name() string {
	return t.t.name
}

func (t *tcpTask) Prepare(env *dvar.EvalEnv) error {
	pr := []uint16{}
	addr := ""

	// 1) check whether has a port_list field, if so use it
	// 2) or check whether env has a port field, if so use it
	if dv, ok := env.Get("target", "port"); ok {
		if port, ok := dv.Port(); !ok {
			return fmt.Errorf("tcp_task(%s), invalid port number", t.name())
		} else {
			pr = append(pr, port)
		}
	} else if dv, ok := env.Get("target", "port_list"); ok {
		if prList, ok := dv.PortList(); !ok {
			return fmt.Errorf("tcp_task(%s), invalid port list", t.name())
		} else {
			pr = prList
		}
	} else {
		return fmt.Errorf(
			"tcp_task(%s) does not have port/port_list definition",
			t.name(),
		)
	}

	// addrese
	if dv, ok := env.Get("target", "addr"); ok {
		addr = dv.String()
	} else if dv, ok := env.Get("target", "ip"); ok {
		addr = dv.String()
	} else {
		return fmt.Errorf(
			"tcp_task(%s) does not have proper address define, either address/ip should be defined",
			t.name(),
		)
	}

	t.address = addr
	t.portRange = pr
	return nil
}

func (t *tcpTask) Description() string {
	return fmt.Sprintf("tcp_task[%s]", t.name())
}

func (t *tcpTask) runTcpTask(env *dvar.EvalEnv, port uint16) *tcpTaskResult {
	addrAndPort := fmt.Sprintf("%s:%d", t.address, port)
	stat := &tcpTaskResult{
		Port:    port,
		Address: t.address,
	}

	start := time.Now().UnixMilli()
	d := net.Dialer{Timeout: time.Duration(t.t.timeout) * time.Second}
	conn, err := d.Dial("tcp", addrAndPort)
	end := time.Now().UnixMilli()
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	stat.Timestamp = start
	stat.RT = (end - start)

	if err != nil {
		stat.Error = fmt.Sprintf("%s", err)
		stat.OK = false
	} else {
		stat.OK = true
		stat.LocalAddress = conn.LocalAddr().String()
	}

	return stat
}

func (t *tcpTask) Run(env *dvar.EvalEnv) error {
	for _, port := range t.portRange {
		// run the tcp task
		stat := util.ToMapInterface(t.runTcpTask(env, port))

		// record the result
		env.RecordHistoricalResult(
			"tcp",
			stat,
		)

		// run the check
		if err := t.t.check.Run(env); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	task.RegisterTaskFactory("tcp", &tcpTaskFactory{})
}
