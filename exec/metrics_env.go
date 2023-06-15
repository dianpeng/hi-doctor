package exec

import (
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/metrics"
)

// Execution part of the metrics, exposed by extension
type counter struct {
	name   string
	client metrics.Client
}

func (c *counter) Emit(x interface{}, opt ...interface{}) error {
	y := make(metrics.Option)

	if len(opt) >= 1 {
		opt0, _ := opt[0].(map[string]interface{})
		y = opt0
	}
	return c.client.Emit(c.name, metrics.MetricsCounter, x, y)
}

func (c *counter) Name() string {
	return c.name
}

func (c *counter) Type() string {
	return "counter"
}

type gauge struct {
	name   string
	client metrics.Client
}

func (c *gauge) Emit(x interface{}, opt ...interface{}) error {
	y := make(metrics.Option)

	if len(opt) >= 1 {
		opt0, _ := opt[0].(map[string]interface{})
		y = opt0
	}
	return c.client.Emit(c.name, metrics.MetricsGauge, x, y)
}

func (c *gauge) Name() string {
	return c.name
}

func (c *gauge) Type() string {
	return "gauge"
}

func addBaseLibraryMetrics(e *Executor, env *dvar.EvalEnv) {
	lib := env.GetNamespace("metrics")
	cli := e.Plan().Metrics
	for _, entry := range e.Plan().MetricsList {
		switch entry.Type {
		case metrics.MetricsCounter:
			lib[entry.Name] = &counter{
				name:   entry.Key,
				client: cli,
			}
			break
		case metrics.MetricsGauge:
			lib[entry.Name] = &gauge{
				name:   entry.Key,
				client: cli,
			}
			break

		default:
			panic("unknown metrics type")
			break
		}
	}
}
