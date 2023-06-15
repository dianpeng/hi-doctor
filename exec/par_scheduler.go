package exec

import (
	"fmt"
	"github.com/alitto/pond"
	"github.com/dianpeng/hi-doctor/dvar"
)

type parScheduler struct {
	maxJob int
}

func (p *parScheduler) getPipelineSize(x ScheduleItemList) int {
	return len(x)
}

func (p *parScheduler) getBatchTaskSize(x ScheduleBatch) int {
	cnt := 0
	for _, x0 := range x.Batch {
		cnt += len(x0.Task)
	}
	return cnt
}

type parCtx struct {
	err   error
	batch ScheduleBatch
}

func (p *parScheduler) Run(
	e *Executor,
	x ScheduleItemList,
	base *dvar.EvalEnv,
	l SchedulerEventListener,
) (outE error) {
	var taskIdx int64 = 0
	var batchIdx int64 = 0
	ctxList := []*parCtx{}

	pool := pond.New(p.maxJob, p.getPipelineSize(x))
	defer func() {
		pool.StopAndWait()
		if e := recover(); e != nil {
			outE = fmt.Errorf("panic: %v", e)
		} else {
			for _, y := range ctxList {
				if y.err != nil {
					outE = y.err // just pick the first one which has an error
				}
			}
		}
	}()

	for _, batch := range x {
		ctx := &parCtx{
			err:   nil,
			batch: batch,
		}
		ctxList = append(ctxList, ctx)

		runBatch := func() {
			env := newEvalEnvFromBase(e, base)
			e.pushCurEnv(env)
			env.Set("task", "batch_index", dvar.NewIntVal(batchIdx))

			// event handling
			if err := l.OnBeforeTaskBatch(env); err != nil {
				e.popCurEnv()
				ctx.err = err
				return
			}

			// the batch execution
			for _, v := range ctx.batch.Batch {
				target := v.TargetItem
				taskList := v.Task

				// make sure the batch run linearly
				for _, task := range taskList {
					env.Set("task", "task_index", dvar.NewIntVal(taskIdx))

					// task execution
					target.SetupEnv(env)
					if err := task.Prepare(env); err != nil {
						e.popCurEnv()
						ctx.err = err
						return
					}
					if err := task.Run(env); err != nil {
						e.popCurEnv()
						ctx.err = err
						return
					}
					target.DelEnv(env)
					taskIdx++
				}
			}

			// event handling
			if err := l.OnAfterTaskBatch(env); err != nil {
				e.popCurEnv()
				ctx.err = err
				return
			}

			e.popCurEnv()
		}

		// submit the task
		pool.Submit(runBatch)

		batchIdx++
		taskIdx += int64(p.getBatchTaskSize(batch))
	}

	return nil
}

func newParScheduler(x int) *parScheduler {
	if x <= 0 {
		x = 1
	}
	return &parScheduler{
		maxJob: x,
	}
}
