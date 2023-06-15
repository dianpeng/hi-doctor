package exec

import (
	"github.com/dianpeng/hi-doctor/dvar"
)

// Extension, for any user that is outside of the builtin. It is allowed to
// register extension library to the runtime
type Extension struct {
	Name    string // name of the extension
	Inline  bool   // whether to inline the extension directly
	Library map[string]interface{}
}

type ExtensionFactory interface {
	Create(e *Executor) Extension
	Description() string
}

var extensionFactory = make(map[string]ExtensionFactory)

func AddExtension(name string, ext ExtensionFactory) {
	extensionFactory[name] = ext
}

func addExtension(e *Executor, env *dvar.EvalEnv) {
	lib := env.GetNamespace("ext")
	for _, x := range extensionFactory {
		ext := x.Create(e)
		if ext.Inline {
			env.ExprEnv()[ext.Name] = ext.Library
		} else {
			lib[ext.Name] = ext.Library
		}
	}
}

/* ---------------------------------------------------------------------------
 * builtin-library
 * --------------------------------------------------------------------------*/

// common library for different phase

func addBaseLibraryLog(e *Executor, env *dvar.EvalEnv) {
	libLog := env.GetNamespace("log")
	{
		// the expr library has an impelementation issue that it always expects
		// a function to return something, even nil. This is not working for
		// log case since log.XXX does not return anything

		libLog["Info"] = func(fmt string, a ...any) error {
			e.Log.Info(fmt, a...)
			return nil
		}

		libLog["Warn"] = func(fmt string, a ...any) error {
			e.Log.Warn(fmt, a...)
			return nil
		}

		libLog["Error"] = func(fmt string, a ...any) error {
			e.Log.Error(fmt, a...)
			return nil
		}
	}
}

func addBaseLibraryVar(e *Executor, env *dvar.EvalEnv) {
	lib := env.GetNamespace("var")
	{
		lib["SetLocal"] = func(name string, v interface{}) bool {
			vv := dvar.NewInterfaceVal(v) // never crash
			e.curEnv().Set("local", name, vv)
			return true
		}
		lib["SetGlobal"] = func(name string, v interface{}) bool {
			vv := dvar.NewInterfaceVal(v) // never crash
			e.curEnv().Set("global", name, vv)
			return true
		}
	}
}

type storageOpt struct {
	ty int
	v  interface{}
}

const (
	storageTypeInt = iota
	storageTypeReal
	storageTypeBool

	// compound type used for certain matrix recording etc ...
	storageTypeStructMapInt
	storageTypeStructMapReal
	storageTypeStructMapBool
)

type storageTypeMapInt struct {
	value map[string]int64
}

type storageTypeMapReal struct {
	value map[string]float64
}

func addStorageLibrary(e *Executor, env *dvar.EvalEnv) {
	lib := env.GetNamespace("storage")

	lib["Int"] = func(v int) *storageOpt {
		return &storageOpt{
			ty: storageTypeInt,
			v:  int64(v),
		}
	}
	lib["Real"] = func(v float64) *storageOpt {
		return &storageOpt{
			ty: storageTypeReal,
			v:  v,
		}
	}
	lib["Boolean"] = func(v bool) *storageOpt {
		return &storageOpt{
			ty: storageTypeBool,
			v:  v,
		}
	}
	lib["MapInt"] = func() *storageOpt {
		return &storageOpt{
			ty: storageTypeStructMapInt,
		}
	}
	lib["MapReal"] = func() *storageOpt {
		return &storageOpt{
			ty: storageTypeStructMapReal,
		}
	}
	lib["MapBoolean"] = func() *storageOpt {
		return &storageOpt{
			ty: storageTypeStructMapBool,
		}
	}
}

const (
	triggerTypeCron = iota
	triggerTypeNow
)

type triggerOpt struct {
	ty   int
	expr string
}

func addTriggerLibrary(e *Executor, env *dvar.EvalEnv) {
	funcMap := env.GetNamespace("trigger")

	funcMap["Cron"] = func(v string) *triggerOpt {
		return &triggerOpt{
			ty:   triggerTypeCron,
			expr: v,
		}
	}

	funcMap["Now"] = func() *triggerOpt {
		return &triggerOpt{
			ty:   triggerTypeNow,
			expr: "",
		}
	}
}

const (
	schedulerTypeSeq = iota
	schedulerTypeParallel
)

type schedulerOpt struct {
	ty       int
	maxCount int
}

func addSchedulerLibrary(env *dvar.EvalEnv) {
	field := env.GetNamespace("scheduler")

	field["Sequence"] = func() *schedulerOpt {
		return &schedulerOpt{
			ty: schedulerTypeSeq,
		}
	}
	field["Parallel"] = func(max int) *schedulerOpt {
		return &schedulerOpt{
			ty:       schedulerTypeParallel,
			maxCount: max,
		}
	}
}

// add those extra information into the map
func addInfo(e *Executor, env *dvar.EvalEnv) {
	info := env.GetNamespace("info")
	info["origin"] = e.p.Info.Origin
}

func addBaseLibrary(e *Executor, env *dvar.EvalEnv) {
	addExtension(e, env)
	addBaseLibraryLog(e, env)
	addBaseLibraryVar(e, env)
	addStorageLibrary(e, env)
	addTriggerLibrary(e, env)
	addSchedulerLibrary(env)
	addBaseLibraryMetrics(e, env)
	addInfo(e, env)
}

// ----------------------------------------------------------------------------
// Interfaces
// ----------------------------------------------------------------------------
func newEvalEnv(e *Executor) *dvar.EvalEnv {
	x := dvar.NewEvalEnv()
	x.InheritInNamespace("assets", e.assets)
	addBaseLibrary(e, x)
	return x
}

func newEvalEnvFromBase(e *Executor, base *dvar.EvalEnv) *dvar.EvalEnv {
	x := dvar.NewEvalEnv()
	x.InheritFromEnv(base)
	return x
}

func newEvalEnvForGuard(e *Executor) *dvar.EvalEnv {
	return newEvalEnv(e)
}

func newEvalEnvForTrigger(e *Executor) *dvar.EvalEnv {
	return newEvalEnv(e)
}

func newEvalEnvForTarget(e *Executor) *dvar.EvalEnv {
	return newEvalEnv(e)
}

func newEvalEnvForActive(e *Executor) *dvar.EvalEnv {
	return newEvalEnv(e)
}
