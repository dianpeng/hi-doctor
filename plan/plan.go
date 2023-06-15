package plan

import (
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/fetch"
	"github.com/dianpeng/hi-doctor/metrics"
	"github.com/dianpeng/hi-doctor/spec"
	"github.com/dianpeng/hi-doctor/task"
	"github.com/dianpeng/hi-doctor/trigger"

	"fmt"
	"time"
)

type varMap map[string]dvar.DVar

type Target struct {
	Fetch  fetch.FetcherFactory
	Format dvar.DVar
	Count  uint
	Inline []map[string]interface{}
}

type TaskPlannerItem struct {
	Guard   dvar.DVar        // guard of this task
	Planner task.TaskPlanner // task planner
}

type TaskPlannerList []TaskPlannerItem

// extra info
type Info struct {
	Origin      string    `json:"origin"`
	Md5Checksum string    `json:"md5"` // md5 checksum
	Timestamp   time.Time `json:"timestamp"`

	source string   `json:"-"`
	Source []string `json:"source"` // We split the yaml source based on the
	// line breaking character and store them
	// into an array
}

type ExecuteInfo struct {
	LastExecute  string `json:"last_execute"`
	LastDuration string `json:"last_duration"`
	LastError    error  `json:"last_error"`
	ExecuteTimes uint64 `json:"execute_times"`

	lastDuration time.Duration
	lastExecute  time.Time
}

type Plan struct {
	Name            string                `json:"name"`    // string of the plan, unqiue
	Comment         string                `json:"comment"` // comment of plan
	Info            Info                  `json:"info"`    // plan information
	LogPrefix       string                `json:"-"`       // trace prefix
	Metrics         metrics.Client        `json:"-"`       // client of the metrices
	MetricsList     []metrics.MetricsItem `json:"-"`       // list of metrics object
	Guard           dvar.DVar             `json:"-"`       // guard of the execution of plan
	Storage         varMap                `json:"-"`       // storage (shared variable)
	Global          varMap                `json:"-"`       // global (global variable)
	Local           varMap                `json:"-"`       // local (local variable)
	Trigger         dvar.DVar             `json:"-"`       // trigger of the plan
	Target          Target                `json:"-"`       // target of the plan
	Scheduler       dvar.DVar             `json:"-"`       // scheduler
	TaskPlannerList TaskPlannerList       `json:"-"`       // list of task planner
	Finally         dvar.CodeBlock        `json:"-"`       // finally block of plan

	// Filled by the runtime
	ExecuteInfo ExecuteInfo `json:"execute_info"`

	isStopped bool
	cronId    trigger.CronId
}

func newPlan() *Plan {
	return &Plan{
		Name:    "",
		Comment: "",
		Storage: make(varMap),
		Global:  make(varMap),
		Local:   make(varMap),
	}
}

func Compile(m *spec.Model) (*Plan, error) {
	c := &compiler{
		output: newPlan(),
		model:  m,
	}

	if err := c.compile(); err != nil {
		return nil, err
	} else {
		return c.output, nil
	}
}

func (p *Plan) Stop() {
	p.isStopped = true
	if p.cronId >= 0 {
		trigger.Remove(p.cronId)
		p.Metrics.Stop()
		p.cronId = -1
	}
}

func (p *Plan) SetCronId(cid trigger.CronId) {
	p.cronId = cid
}

func (e *ExecuteInfo) SetLastExecute(l time.Time) {
	e.lastExecute = l
	e.LastExecute = l.String()
}

func (e *ExecuteInfo) SetLastDuration(d time.Duration) {
	e.lastDuration = d
	e.LastDuration = d.String()
}

func (e *ExecuteInfo) IncExecuteTimes() {
	e.ExecuteTimes = e.ExecuteTimes + 1
}

func (e *ExecuteInfo) Description() string {
	return fmt.Sprintf("start_ts: %s; duration: %s;",
		e.LastExecute,
		e.LastDuration,
	)
}
