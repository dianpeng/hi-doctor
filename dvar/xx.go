package dvar

func must(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}

func unreachable(msg string) {
	panic(msg)
}
