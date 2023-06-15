package builtin

import (
	"fmt"

	"github.com/dianpeng/hi-doctor/check"
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/spec"
	"github.com/dianpeng/hi-doctor/task"

	"github.com/mitchellh/mapstructure"
)

// Basic code skeleton, do nothing special but just allow user to run some
// abitary code as they wish

type codeTask struct {
	t *codeTaskTemplate
}

type codeTaskTemplate struct {
	code  dvar.CodeBlock
	check check.Check
}

type codeTaskFactory struct {
}

type codeBlockRaw struct {
	CodeBlock []string `mapstructure:"code_block"`
}

func (c *codeTaskTemplate) Description() string {
	return "code"
}

func (c *codeTaskTemplate) GenTask(env *dvar.EvalEnv) (task.TaskList, error) {
	xx := &codeTask{
		t: c,
	}

	return task.TaskList{
		xx,
	}, nil
}

func (c *codeTask) Prepare(env *dvar.EvalEnv) error {
	return nil
}

func (c *codeTask) Run(env *dvar.EvalEnv) error {
	for i, code := range c.t.code {
		if _, err := code.Value(env); err != nil {
			return fmt.Errorf("code task statement[%d] execution error %s", i, err)
		}
	}
	if err := c.t.check.Run(env); err != nil {
		return err
	}
	return nil
}

func (c *codeTask) Description() string {
	return "code"
}

func (c *codeTaskFactory) SanityCheck(x spec.TaskOption) error {
	out := &codeBlockRaw{}
	err := mapstructure.Decode(x, out)
	if err != nil {
		return fmt.Errorf("code task, invalid option: %s", err)
	}
	return nil
}

func (c *codeTaskFactory) Compile(x spec.TaskOption, checkSpec *spec.Check) (task.TaskPlanner, error) {
	out := &codeTaskTemplate{}

	if ck, err := check.CompileCheck(checkSpec); err != nil {
		return nil, err
	} else {
		out.check = ck
	}

	cbr := &codeBlockRaw{}
	if err := mapstructure.Decode(x, cbr); err != nil {
		return nil, fmt.Errorf("code task, invalid option: %s", err)
	} else {
		for idx, code := range cbr.CodeBlock {
			if dv, err := dvar.NewDVarScriptContext(code); err != nil {
				return nil, fmt.Errorf("code task's code_block[%d] compilation error %s", idx, err)
			} else {
				out.code = append(out.code, dv)
			}
		}
	}

	return out, nil
}

func init() {
	task.RegisterTaskFactory("code", &codeTaskFactory{})
}
