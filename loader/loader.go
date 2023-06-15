package loader

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/dianpeng/hi-doctor/spec"
	"github.com/dianpeng/hi-doctor/task"
)

func errMissingField(field string) error {
	return fmt.Errorf("%s is missing", field)
}

func checkTask(idx int, t *spec.TaskAny) error {
	taskFactory := task.GetTaskFactory(t.Type) // type of the task
	if taskFactory == nil {
		return fmt.Errorf("inspection.task[%d].type %s is unknown to us", idx, t.Type)
	}

	if err := taskFactory.SanityCheck(t.Option); err != nil {
		return err
	}

	return nil
}

func sanityCheck(ins *spec.Model) error {
	// metadata
	if ins.Name == "" {
		return errMissingField("inspection.name")
	}

	// target
	if ins.Target == nil {
		return errMissingField("inspection.target")
	}

	// trigger
	if ins.Trigger == "" {
		return errMissingField("inspection.trigger")
	}

	// task
	for idx, v := range ins.Task {
		if err := checkTask(idx, v); err != nil {
			return err
		}
	}

	return nil
}

func ParseData(data string) (*spec.Model, error) {
	ins := &spec.Model{}

	if err := yaml.Unmarshal([]byte(data), ins); err != nil {
		return nil, err
	}
	if err := sanityCheck(ins); err != nil {
		return nil, err
	}

	return ins, nil
}
