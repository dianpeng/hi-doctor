package task

import (
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/spec"
)

type Task interface {
	// Run this phase's preparation phase. This is a logic step and mostly each
	// phase should not be awared of this
	Prepare(*dvar.EvalEnv) error

	// For each phase, all the observable effects will be stored by global variable
	// inside of the EvalEnv. User is expected to know this and we do not allow
	// customized binding, ie local variables
	Run(*dvar.EvalEnv) error

	// description of this Task
	Description() string
}

type TaskList []Task

// Each Task definition will result in a TaskPlanner interface which shows us
// how to generate a Task list along with its execution
type TaskPlanner interface {
	Description() string

	// generate a list of Task
	GenTask(*dvar.EvalEnv) (TaskList, error)
}

type TaskFactory interface {
	// Sanity check task option is correct or not
	SanityCheck(spec.TaskOption) error

	// Compile the task option
	Compile(spec.TaskOption, *spec.Check) (TaskPlanner, error)
}
