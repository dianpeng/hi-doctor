package task

var (
	factoryMap = make(map[string]TaskFactory)
)

func GetTaskFactory(name string) TaskFactory {
	out, ok := factoryMap[name]
	if !ok {
		return nil
	}
	return out
}

func RegisterTaskFactory(name string, f TaskFactory) {
	_, ok := factoryMap[name]
	if ok {
		panic("factory already register")
	}

	factoryMap[name] = f
}
