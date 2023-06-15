package metrics

import (
	"fmt"
	"net/http"

	"github.com/dianpeng/hi-doctor/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var register = prometheus.NewRegistry()

type promMetrics struct {
	key string
	ty  int
	itf interface{}
}

type promClient struct {
	prefix  string // prefix of the client, will be inserted into the tag
	metrics map[string]promMetrics
}

func (p *promClient) toTag(x Option) prometheus.Labels {
	out := make(prometheus.Labels)
	if x == nil {
		return out
	}
	for k, v := range x {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

func (p *promClient) Stop() {
	for _, v := range p.metrics {
		register.Unregister(v.itf.(prometheus.Collector))
	}
}

func (p *promClient) Define(key string, ty int, tag Option) error {
	var itf interface{}

	switch ty {
	case MetricsCounter:
		v := prometheus.NewCounter(prometheus.CounterOpts{
			Name:        key,
			Help:        fmt.Sprintf("metrics client[%s] key %s", p.prefix, key),
			ConstLabels: p.toTag(tag),
		})
		if err := register.Register(v); err != nil {
			return fmt.Errorf("metrics client[%s] key %s cannot register", p.prefix, key)
		}
		itf = v
		break

	case MetricsGauge:
		v := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        key,
			Help:        fmt.Sprintf("metrics client[%s] key %s", p.prefix, key),
			ConstLabels: p.toTag(tag),
		})
		if err := register.Register(v); err != nil {
			return fmt.Errorf("metrics client[%s] key %s cannot register", p.prefix, key)
		}
		itf = v
		break
	}

	p.metrics[key] = promMetrics{
		key: key,
		ty:  ty,
		itf: itf,
	}
	return nil
}

func (p *promClient) Emit(
	key string,
	ty int,
	value interface{},
	tag Option,
) error {
	vv, ok := p.metrics[key]
	if !ok {
		return fmt.Errorf("metrics %s not found", key)
	}
	if vv.ty != ty {
		return fmt.Errorf("metrics %s type mismatch", key)
	}

	switch ty {
	case MetricsCounter:
		if value, ok := util.ToReal(value, true); !ok {
			return fmt.Errorf("metrics %s with value %v is invalid", key, value)
		} else {
			c := vv.itf.(prometheus.Counter)
			c.Add(value)
			return nil
		}

	case MetricsGauge:
		if value, ok := util.ToReal(value, true); !ok {
			return fmt.Errorf("metrics %s with value %v is invalid", key, value)
		} else {
			c := vv.itf.(prometheus.Gauge)
			c.Set(float64(value))
			return nil
		}

	default:
		panic("unknown type")
		return nil
	}
}

type promFactory struct{}

func (p *promFactory) Create(
	prefix string,
	opt Option,
) (Client, error) {
	return &promClient{
		prefix:  prefix,
		metrics: make(map[string]promMetrics),
	}, nil
}

func PrometheusHttpHandler() http.Handler {
	return promhttp.HandlerFor(
		register,
		promhttp.HandlerOpts{Registry: register},
	)
}

func init() {
	AddClientFactory("prometheus", &promFactory{})
}
