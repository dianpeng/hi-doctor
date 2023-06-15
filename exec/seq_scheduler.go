package exec

import (
	"github.com/dianpeng/hi-doctor/dvar"
)

// basic scheduler
type seqScheduler struct {
}

func (_ *seqScheduler) Run(
	e *Executor,
	x ScheduleItemList,
	base *dvar.EvalEnv,
	l SchedulerEventListener,
) error {

	var taskIdx int64 = 0
	var batchIdx int64 = 0

	for _, batch := range x {
		env := newEvalEnvFromBase(e, base)
		e.pushCurEnv(env)
		env.Set("task", "batch_index", dvar.NewIntVal(batchIdx))

		// event handling
		if err := l.OnBeforeTaskBatch(env); err != nil {
			e.popCurEnv()
			return err
		}

		// the batch execution
		for _, v := range batch.Batch {
			target := v.TargetItem
			taskList := v.Task

			// make sure the batch run linearly
			for _, task := range taskList {
				env.Set("task", "task_index", dvar.NewIntVal(taskIdx))

				// task execution
				target.SetupEnv(env)
				if err := task.Prepare(env); err != nil {
					e.popCurEnv()
					return err
				}
				if err := task.Run(env); err != nil {
					e.popCurEnv()
					return err
				}
				target.DelEnv(env)
				taskIdx++
			}
		}

		// event handling
		if err := l.OnAfterTaskBatch(env); err != nil {
			e.popCurEnv()
			return err
		}

		batchIdx++
		e.popCurEnv()
	}

	return nil
}

func newSeqScheduler() *seqScheduler {
	return &seqScheduler{}
}
