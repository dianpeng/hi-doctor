package metrics

const (
	MetricsCounter = iota
	MetricsGauge
)

type Option map[string]interface{}

type MetricsItem struct {
	Name string
	Key  string
	Type int
}

type Client interface {
	Define(string, int, Option) error
	Emit(string, int, interface{}, Option) error
	Stop()
}

type ClientFactory interface {
	Create(string, Option) (Client, error)
}

// Registery ------------------------------------------------------------------
var (
	reg = make(map[string]ClientFactory)
)

func AddClientFactory(
	name string,
	f ClientFactory,
) {
	reg[name] = f
}

func GetClientFactory(name string) ClientFactory {
	f, ok := reg[name]
	if !ok {
		return nil
	}
	return f
}
