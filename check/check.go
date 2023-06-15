package check

import (
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/spec"

	"fmt"
)

type Check struct {
	Condition dvar.DVar
	Then      dvar.CodeBlock
	Otherwise dvar.CodeBlock
	Lastly    dvar.CodeBlock
}

func newNullCheck() Check {
	return Check{
		Condition: dvar.NewDVarLit(""),
	}
}

func CompileCheck(m *spec.Check) (Check, error) {
	ck := newNullCheck()
	if m == nil {
		return ck, nil
	}

	if dv, err := dvar.NewDVarScriptContext(m.Condition); err != nil {
		return Check{}, fmt.Errorf("check.Condition compile failed: %s", err)
	} else {
		ck.Condition = dv
	}

	if v, err := dvar.CompileCodeBlock("check.Then", m.Then); err != nil {
		return Check{}, err
	} else {
		ck.Then = v
	}
	if v, err := dvar.CompileCodeBlock("check.Otherwise", m.Otherwise); err != nil {
		return Check{}, err
	} else {
		ck.Otherwise = v
	}
	if v, err := dvar.CompileCodeBlock("check.Lastly", m.Lastly); err != nil {
		return Check{}, err
	} else {
		ck.Lastly = v
	}

	return ck, nil
}

func (c *Check) runCodeBlock(
	context string,
	x dvar.CodeBlock,
	env *dvar.EvalEnv,
) error {
	for idx, xx := range x {
		if _, err := xx.Value(env); err != nil {
			return fmt.Errorf(
				"check.%s[%d] execution failed: %s",
				context,
				idx,
				err,
			)
		}
	}
	return nil
}

func (c *Check) Run(env *dvar.EvalEnv) error {
	output, err := c.Condition.Value(env)
	if err != nil {
		return fmt.Errorf("check.condition execution failed: %s", err)
	}
	env.Set("check", "condition", output)

	// if the condition failed, then run down each otherwise step until we are
	// done
	if output.Boolean() {
		if err := c.runCodeBlock(
			"then",
			c.Then,
			env,
		); err != nil {
			return err
		}
	} else {
		if err := c.runCodeBlock(
			"otherwise",
			c.Otherwise,
			env,
		); err != nil {
			return err
		}
	}

	return c.runCodeBlock("lastly", c.Lastly, env)
}
