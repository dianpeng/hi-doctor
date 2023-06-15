package try

// Try defines a general way for trying certain inspection task.

const (
	// Exact try, ie specify retry # of times and execute that *task*
	// accordingly. This is the most basic way.
	TryExact = iota

	// Try at most the specified times, and additionally specify the
	// bailout condition, ie when the # of this threshold reached, the
	// try procedure just aborts
	TryAtMost
)

type TryOpt struct {
	Type      int
	TryCount  int
	Threshold int
}
