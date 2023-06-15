package plan

import (
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/fetch"
	"github.com/dianpeng/hi-doctor/metrics"
	"github.com/dianpeng/hi-doctor/spec"
	"github.com/dianpeng/hi-doctor/task"

	"fmt"
	"strings"
)

// compile a spec model to an internal representation, ie plan object
type compiler struct {
	output *Plan
	model  *spec.Model
}

func (c *compiler) compileDVarMap(
	context string,
	out varMap,
	kv map[string]string,
) error {
	for k, v := range kv {
		dv, err := dvar.NewDVarScriptContext(v)
		if err != nil {
			return fmt.Errorf("%s[%s] compile failed: %s", context, k, err)
		}
		out[k] = dv
	}
	return nil
}

func (c *compiler) compileMetrics(
	m *spec.Metrics, // metrics model
) (metrics.Client, []metrics.MetricsItem, error) {
	if m == nil {
		return nil, nil, nil // no-op, user do not need metrics
	}

	factory := metrics.GetClientFactory(m.Provider)
	if factory == nil {
		return nil, nil, fmt.Errorf("metrics factory %s is unknown to us", m.Provider)
	}

	cli, err := factory.Create(m.Namespace, m.Option)
	if err != nil {
		return nil, nil, err
	}

	mlist := []metrics.MetricsItem{}

	for idx, entry := range m.Define {
		ty := 0
		switch entry.Type {
		case "counter":
			ty = metrics.MetricsCounter
			break

		case "gauge":
			ty = metrics.MetricsGauge
			break

		default:
			return nil, nil, fmt.Errorf("metrics[%d] type %s is unknown", idx, entry.Type)
		}

		if err := cli.Define(entry.Key, ty, metrics.Option(entry.Label)); err != nil {
			return nil, nil, fmt.Errorf("metrics[%d] define fail %s", idx, err)
		}

		mlist = append(mlist, metrics.MetricsItem{
			Name: entry.Name,
			Type: ty,
			Key:  entry.Key,
		})
	}

	return cli, mlist, nil
}

// ----------------------------------------------------------------------------
// Guard

func (c *compiler) compileGuard() error {
	if dv, err := dvar.NewDVarScriptContext(c.model.Guard); err != nil {
		return err
	} else {
		c.output.Guard = dv
		return nil
	}
}

// ----------------------------------------------------------------------------
// Storage compilation

func (c *compiler) compileStorage() error {
	return c.compileDVarMap(
		"storage",
		c.output.Storage,
		c.model.Storage,
	)
}

// ----------------------------------------------------------------------------
// Global compilation
func (c *compiler) compileGlobal() error {
	return c.compileDVarMap(
		"global",
		c.output.Global,
		c.model.Global,
	)
}

// ----------------------------------------------------------------------------
// Local compilation
func (c *compiler) compileLocal() error {
	return c.compileDVarMap(
		"local",
		c.output.Local,
		c.model.Local,
	)
}

// ----------------------------------------------------------------------------
// Trigger compilation
func (c *compiler) compileTrigger() error {
	dv, err := dvar.NewDVarScriptContext(c.model.Trigger)
	if err != nil {
		return fmt.Errorf("trigger compile failed: %s", err)
	}
	c.output.Trigger = dv
	return nil
}

// ----------------------------------------------------------------------------
// Target compilation
func (c *compiler) compileTarget() error {
	if c.model.Target == nil {
		return fmt.Errorf("target is not specified")
	}

	dv, err := dvar.NewDVarStringContext(c.model.Target.Format)
	if err != nil {
		return fmt.Errorf("target.format compile failed: %s", err)
	}

	if c.model.Target.Fetch != nil {
		fetch, err := fetch.Compile(c.model.Target.Fetch)
		if err != nil {
			return fmt.Errorf("target.fetch compile failed: %s", err)
		}
		c.output.Target.Fetch = fetch
		c.output.Target.Format = dv
	} else if c.model.Target.Inline != nil {
		c.output.Target.Inline = c.model.Target.Inline
	} else if c.model.Target.Count > 0 {
		c.output.Target.Count = c.model.Target.Count
	} else {
		return fmt.Errorf("target is invalid, at least specify Fetch/Inline/Count")
	}
	return nil
}

// ----------------------------------------------------------------------------
// Schedule
func (c *compiler) compileScheduler() error {
	if dv, err := dvar.NewDVarScriptContext(c.model.Scheduler); err != nil {
		return err
	} else {
		c.output.Scheduler = dv
		return nil
	}
}

// ----------------------------------------------------------------------------
// TaskList
func (c *compiler) compileTaskList() error {
	for i, tany := range c.model.Task {
		factory := task.GetTaskFactory(tany.Type)
		if factory == nil {
			return fmt.Errorf("task[%d] type %s unknown to us", i, tany.Type)
		}
		taskPlanner, err := factory.Compile(tany.Option, tany.Check)
		if err != nil {
			return fmt.Errorf("task[%d(%s)] cannot be created: %s", i, tany.Type, err)
		}

		guard, err := dvar.NewDVarScriptContext(tany.Guard)
		if err != nil {
			return fmt.Errorf("task[%d(%s)].guard cannot be created: %s", i, tany.Type, err)
		}

		c.output.TaskPlannerList = append(c.output.TaskPlannerList, TaskPlannerItem{
			Guard:   guard,
			Planner: taskPlanner,
		})
	}
	return nil
}

// ----------------------------------------------------------------------------
// Finally
func (c *compiler) compileFinally() error {
	for i, f := range c.model.Finally {
		if dv, err := dvar.NewDVarScriptContext(f); err != nil {
			return fmt.Errorf("finally[%d] compile to fail: %s", i, err)
		} else {
			c.output.Finally = append(c.output.Finally, dv)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Entry
func (c *compiler) compile() error {
	c.output.Name = c.model.Name
	c.output.Comment = c.model.Comment
	c.output.LogPrefix = c.model.LogPrefix

	c.output.Info.Origin = c.model.Info.Origin
	c.output.Info.Md5Checksum = c.model.Info.Md5
	c.output.Info.Timestamp = c.model.Info.Timestamp

	slist := strings.Split(c.model.Info.Source, "\n")
	c.output.Info.Source = slist
	c.output.Info.source = c.model.Info.Source

	if a, b, err := c.compileMetrics(c.model.Metrics); err != nil {
		return err
	} else {
		c.output.Metrics = a
		c.output.MetricsList = b
	}

	if err := c.compileGuard(); err != nil {
		return err
	}

	if err := c.compileStorage(); err != nil {
		return err
	}

	if err := c.compileGlobal(); err != nil {
		return err
	}

	if err := c.compileLocal(); err != nil {
		return err
	}

	if err := c.compileTarget(); err != nil {
		return err
	}

	if err := c.compileTrigger(); err != nil {
		return err
	}

	if err := c.compileScheduler(); err != nil {
		return err
	}

	if err := c.compileTaskList(); err != nil {
		return err
	}

	if err := c.compileFinally(); err != nil {
		return err
	}

	return nil
}
