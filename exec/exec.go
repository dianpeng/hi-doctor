package exec

import (
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/plan"
	"github.com/dianpeng/hi-doctor/storage"
	"github.com/dianpeng/hi-doctor/trace"
	"github.com/dianpeng/hi-doctor/trigger"

	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ----------------------------------------------------------------------------
// execution of the plan.
//
// this place contains all the code for executing a generated plan, ie after
// compilation. Each plan execution is based on a predefined order and operation
//
// The order is been grouped into following phases
//
//  1) Guard Phase
//
//     Evaluate the toggle section to decide whether to register this inspection
//     task or not.
//
//  2) Trigger Phase
//
//     Evaluate the trigger section and register the inspection job trigger
//
//  3) Active Phase
//
//     When the trigger is fired, the Active Phase is entered. During the
//     active phase, each inspection plan is been evaluated as following
//
//    3.1) Target Phase
//
//     Evaluate the target and filter out list of inspection target to work
//     with
//
//    3.2) Task
//
//     For each target in the target list, its taks will be issued
//
// ----------------------------------------------------------------------------

type Executor struct {
	p        *plan.Plan
	assets   dvar.ValMap
	storage  map[string]storage.Storage
	curE     []*dvar.EvalEnv
	runMutex sync.Mutex

	// Opaque structure for any extension to be used
	Blackboard map[string]interface{}
	Log        trace.Trace
}

func (e *Executor) Plan() *plan.Plan {
	return e.p
}

// ----------------------------------------------------------------------------
// during the lifetime, we need to modify the eval environment multiple times
func (e *Executor) curEnv() *dvar.EvalEnv {
	return e.curE[len(e.curE)-1]
}

func (e *Executor) pushCurEnv(env *dvar.EvalEnv) {
	e.curE = append(e.curE, env)
}

func (e *Executor) popCurEnv() {
	e.curE = e.curE[0 : len(e.curE)-1]
}

// ----------------------------------------------------------------------------
func (e *Executor) runGuard() (bool, error) {
	env := newEvalEnvForGuard(e)
	e.pushCurEnv(env)
	defer e.popCurEnv()

	out, err := e.p.Guard.Value(env)
	if err != nil {
		return false, err
	}
	return out.Boolean(), nil
}

// ----------------------------------------------------------------------------
// storage run

func (e *Executor) runStorage(env *dvar.EvalEnv) error {
	getStorageOpt := func(v dvar.Val) *storageOpt {
		if v.IsAny() {
			vv, ok := v.GetAny().(*storageOpt)
			if ok {
				return vv
			}
		}
		return nil
	}

	// evaluate the storage run initializers
	for k, v := range e.p.Storage {
		if val, err := v.Value(env); err != nil {
			return fmt.Errorf(
				"executor.storage(%s) initialization failed: %s",
				k,
				err,
			)
		} else {
			vv := getStorageOpt(val)
			if vv == nil {
				return fmt.Errorf("executor.storage(%s) invalid storage type", k)
			}

			var value storage.Storage

			switch vv.ty {
			case storageTypeInt:
				value = storage.NewInt(vv.v.(int64))
				break

			case storageTypeReal:
				value = storage.NewReal(vv.v.(float64))
				break

			case storageTypeBool:
				value = storage.NewBoolean(vv.v.(bool))
				break

			case storageTypeStructMapInt:
				value = storage.NewMapInt()
				break

			case storageTypeStructMapReal:
				value = storage.NewMapReal()
				break

			case storageTypeStructMapBool:
				value = storage.NewMapBoolean()
				break

			default:
				panic("unknown storage type")
			}
			e.storage[k] = value
		}
	}

	return nil
}

func (e *Executor) defineStorage(env *dvar.EvalEnv) error {
	ns := env.GetNamespace("storage")
	for k, v := range e.storage {
		ns[k] = v
	}
	return nil
}

// ----------------------------------------------------------------------------
// trigger run
//
//  for simplicity we just support 2 types of trigger expression for now,
//  one is cron, and the other is immediate

func (e *Executor) setupTrigger(env *dvar.EvalEnv) error {
	getTriggerType := func(v dvar.Val) *triggerOpt {
		if v.IsAny() {
			vv, ok := v.GetAny().(*triggerOpt)
			if ok {
				return vv
			}
		}
		return nil
	}

	output, err := e.p.Trigger.Value(env)
	if err != nil {
		return err
	}

	vv := getTriggerType(output)
	if vv == nil {
		return fmt.Errorf("executor.trigger invalid run type")
	}

	switch vv.ty {
	case triggerTypeCron:
		if tid, err := trigger.Cron(vv.expr, func() {
			e.runActive()
		}); err != nil {
			return fmt.Errorf("executor.trigger invalid cron: %s", err)
		} else {
			e.p.SetCronId(tid)
		}
		break

	case triggerTypeNow:
		if err := trigger.Now(func() {
			e.runActive()
		}); err != nil {
			return fmt.Errorf("executor.trigger invalid now: %s", err)
		} else {
			e.p.SetCronId(-1)
		}
		break

	default:
		break
	}

	return nil
}

func (e *Executor) runTrigger() error {
	env := newEvalEnvForTrigger(e)
	e.pushCurEnv(env)
	defer e.popCurEnv()

	// (0) run the storage before running the trigger
	if err := e.runStorage(env); err != nil {
		return err
	}

	// (1) evaluate the trigger expression
	return e.setupTrigger(env)
}

// ---------------------------------------------------------------------------
//
//	active phase
func (e *Executor) runVarMap(field string,
	input map[string]dvar.DVar,
	env *dvar.EvalEnv,
) error {
	for k, v := range input {
		if output, err := v.Value(env); err != nil {
			return fmt.Errorf("executor.%s(%s) execution failed: %s", field, k, err)
		} else {
			env.Set(field, k, output)
		}
	}
	return nil
}

func (e *Executor) runLocal(env *dvar.EvalEnv) error {
	return e.runVarMap("local", e.p.Local, env)
}

func (e *Executor) runGlobal(env *dvar.EvalEnv) error {
	return e.runVarMap("global", e.p.Global, env)
}

func (e *Executor) populateTaskList(
	env *dvar.EvalEnv,
	insTarget InspectionTarget,
) (ScheduleItemList, error) {
	tlist := ScheduleItemList{}
	for _, v := range insTarget {
		batch := ScheduleBatch{}

		v.SetupEnv(env)
		for _, pi := range e.p.TaskPlannerList {
			p := pi.Planner

			// evaluate the guard
			if guard, err := pi.Guard.Value(env); err != nil {
				return nil, fmt.Errorf(
					"executor.target task planner(%s) guard failed: %s",
					p.Description(),
					err,
				)
			} else if !guard.Boolean() {
				continue // do nothing, just skip this task
			}

			t, err := p.GenTask(env)
			if err != nil {
				return nil, fmt.Errorf(
					"executor.target task planner(%s) failed: %s",
					p.Description(),
					err,
				)
			}

			batch.Batch = append(batch.Batch, ScheduleItem{
				Guard:      e.p.Guard,
				TargetItem: v,
				Task:       t,
			})
		}
		v.DelEnv(env)

		tlist = append(tlist, batch)
	}

	return tlist, nil
}

func (e *Executor) runTargetInspectionOnRawData(
	env *dvar.EvalEnv,
	rawData string,
) (ScheduleItemList, error) {

	// now run the filter operation, the output will be a valid inspection target
	// object, otherwise it is an error
	output, err := e.p.Target.Format.Value(env)
	if err != nil {
		return nil, fmt.Errorf("executor.target filter operation failed: %s", err)
	}

	insTarget := getInspectionTarget(
		rawData,
		output.String(),
	)

	if insTarget == nil {
		return nil, fmt.Errorf(
			"executor.target format %s invalid, either unknown format or the input "+
				"data, specified at target.raw, cannot be parsed as specified format",
			output.String(),
		)
	}

	return e.populateTaskList(env, insTarget)
}

func (e *Executor) runTargetFetch(baseEnv *dvar.EvalEnv) (ScheduleItemList, error) {
	env := newEvalEnvForTarget(e)
	env.InheritFromEnv(baseEnv)

	fetcher, err := e.p.Target.Fetch.Create(env)
	if err != nil {
		return nil, fmt.Errorf("executor.target create fetcher failed: %s", err)
	}

	data, err := fetcher.Obtain()
	if err != nil {
		return nil, fmt.Errorf("executor.target fetch obtain failed: %s", err)
	}

	// the data must be ETLed into valid format, for now, the data will be
	// just put into string format and feed back to the environment
	env.Set("target", "raw", dvar.NewStringVal(string(data)))

	return e.runTargetInspectionOnRawData(
		env,
		string(data),
	)
}

func (e *Executor) runTargetInline(baseEnv *dvar.EvalEnv) (ScheduleItemList, error) {
	env := newEvalEnvForTarget(e)
	env.InheritFromEnv(baseEnv)

	jsonData, err := json.Marshal(e.p.Target.Inline)
	if err != nil {
		return nil, fmt.Errorf("executor.target.inline is invalid: %s", err)
	}
	return e.runTargetInspectionOnRawData(
		env,
		string(jsonData),
	)
}

func (e *Executor) runTargetCount(baseEnv *dvar.EvalEnv) (ScheduleItemList, error) {
	env := newEvalEnvForTarget(e)
	env.InheritFromEnv(baseEnv)
	insTarget := []*InspectionTargetItem{}

	for i := uint(0); i < e.p.Target.Count; i++ {
		insTarget = append(insTarget, &InspectionTargetItem{
			Name:     fmt.Sprintf("Count(%d)", i),
			AnyOther: make(map[string]interface{}),
		})
	}
	return e.populateTaskList(env, insTarget)
}

func (e *Executor) runTarget(baseEnv *dvar.EvalEnv) (ScheduleItemList, error) {
	if e.p.Target.Fetch != nil {
		return e.runTargetFetch(baseEnv)
	}
	if e.p.Target.Inline != nil {
		return e.runTargetInline(baseEnv)
	}
	if e.p.Target.Count > 0 {
		return e.runTargetCount(baseEnv)
	}

	return nil, fmt.Errorf(
		"executor.target invalid, Fetch/Inline/Count must be specified",
	)
}

func (e *Executor) runScheduler(env *dvar.EvalEnv) (Scheduler, error) {
	if dv, err := e.p.Scheduler.Value(env); err != nil {
		return nil, err
	} else {
		if dv.IsAny() {
			if vv, ok := dv.GetAny().(*schedulerOpt); ok {
				switch vv.ty {
				case schedulerTypeParallel:
					return newParScheduler(vv.maxCount), nil
				default:
					return newSeqScheduler(), nil
				}
			}
		}
		return newSeqScheduler(), nil
	}
}

func (e *Executor) runFinally(env *dvar.EvalEnv) error {
	for i, v := range e.p.Finally {
		if _, err := v.Value(env); err != nil {
			return fmt.Errorf("finally[%d] execution failed: %s", i, err)
		}
	}
	return nil
}

func (e *Executor) OnBeforeTaskBatch(env *dvar.EvalEnv) error {
	// setup the local variable, used by
	return e.runLocal(env)
}

func (e *Executor) OnAfterTaskBatch(env *dvar.EvalEnv) error {
	// do nothing
	return nil
}

func (e *Executor) doRunActive() error {
	env := newEvalEnvForActive(e)
	env.InheritInNamespace("assets", e.assets)

	e.pushCurEnv(env)
	defer e.popCurEnv()

	// 0) define all the storage
	if err := e.defineStorage(env); err != nil {
		return err
	}

	// 1) setup the global variable, shared by *all* tasks
	if err := e.runGlobal(env); err != nil {
		return err
	}

	// 2) select scheduler
	scheduler, err := e.runScheduler(env)
	if err != nil {
		return err
	}

	// 3) run the target and generate probing target
	tlist, err := e.runTarget(env)
	if err != nil {
		return err
	}

	// 4) run the probing task
	if err := scheduler.Run(e, tlist, env, e); err != nil {
		return err
	}

	// 5) run the finally code block
	if err := e.runFinally(env); err != nil {
		return err
	}

	return nil
}

func (e *Executor) runActive() {
	// using mutex to avoid been executed multiple times because of cron schedule
	e.runMutex.Lock()
	defer e.runMutex.Unlock()

	e.Log.Info("trigger fired, job start to execute")

	start := time.Now()
	err := e.doRunActive()
	done := time.Now()

	e.p.ExecuteInfo.SetLastDuration(done.Sub(start))
	e.p.ExecuteInfo.SetLastExecute(start)
	e.p.ExecuteInfo.IncExecuteTimes()
	e.p.ExecuteInfo.LastError = err

	if err != nil {
		e.Log.Error(
			"job finish its execution with error status: %s, other info: %s",
			err,
			e.p.ExecuteInfo.Description(),
		)
	} else {
		e.Log.Info("job finish its execution successfully, othter info: %s",
			e.p.ExecuteInfo.Description(),
		)
	}
}

func (e *Executor) TracePrefix() string {
	return fmt.Sprintf(
		"%s{name: %s, origin: %s}",
		e.p.LogPrefix,
		e.p.Name,
		e.p.Info.Origin,
	)
}

func (e *Executor) Start() error {
	if toggle, err := e.runGuard(); err != nil {
		return err
	} else if !toggle {
		return nil // shortcut
	}

	return e.runTrigger()
}

func NewExecutor(assets dvar.ValMap, p *plan.Plan) *Executor {
	exec := &Executor{
		p:          p,
		assets:     assets,
		storage:    make(map[string]storage.Storage),
		Blackboard: make(map[string]interface{}),
	}
	exec.Log = trace.NewTrace(exec)
	return exec
}
