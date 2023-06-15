package exec

import (
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/task"
)

type ScheduleItem struct {
	Guard      dvar.DVar             // guard of task
	TargetItem *InspectionTargetItem // task information
	Task       []task.Task           // task list itself, *must* be run in seq
}

// ScheduleBatch make sure everything inside of will be executed linearly,
// regardlessly of what scheduler is been configured
type ScheduleBatch struct {
	Batch []ScheduleItem
}

type ScheduleItemList []ScheduleBatch

type SchedulerEventListener interface {
	OnBeforeTaskBatch(*dvar.EvalEnv) error
	OnAfterTaskBatch(*dvar.EvalEnv) error
}

type Scheduler interface {
	Run(*Executor, ScheduleItemList, *dvar.EvalEnv, SchedulerEventListener) error
}
