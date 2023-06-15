package assert

import (
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/exec"
)

type assertFactory struct {
}

type AssertInfo struct {
	YesFail int
	NoFail  int
}

func (a *AssertInfo) IsOK() bool {
	return a.YesFail == 0 && a.NoFail == 0
}

func (a *assertFactory) Create(e *exec.Executor) exec.Extension {
	b := e.Blackboard
	info := &AssertInfo{}
	b["assert"] = info

	lib := make(map[string]interface{})

	lib["Yes"] = func(x interface{}) bool {
		v := dvar.NewInterfaceVal(x)
		if !v.Boolean() {
			info.YesFail++
			return false
		}
		return true
	}

	lib["No"] = func(x interface{}) bool {
		v := dvar.NewInterfaceVal(x)
		if v.Boolean() {
			info.YesFail++
			return false
		}
		return true
	}

	lib["OK"] = func() bool {
		return info.IsOK()
	}

	return exec.Extension{
		Name:    "assert",
		Inline:  true,
		Library: lib,
	}
}

func (a *assertFactory) Description() string {
	return "assert"
}

func init() {
	exec.AddExtension("assert", &assertFactory{})
}
