package metrics

import (
	"fmt"
	"log"
	"strings"
)

// local metrics client, ie printing to the stdout
type localMetricsClient struct {
	prefix     string
	constLabel map[string]Option
}

func (l *localMetricsClient) Define(name string, _ int, opt Option) error {
	if opt == nil {
		opt = make(Option)
	}
	l.constLabel[name] = opt
	return nil
}

func (l *localMetricsClient) Stop() {}

func (l *localMetricsClient) Emit(
	key string,
	ty int,
	value interface{},
	tag Option,
) error {
	constL, ok := l.constLabel[key]
	if !ok {
		return fmt.Errorf("metrics %s is not defined", key)
	}

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("[%s; %s]", l.prefix, key))

	sb.WriteString("{")
	for k, v := range constL {
		sb.WriteString(fmt.Sprintf("%s=%v;", k, v))
	}
	for k, v := range tag {
		sb.WriteString(fmt.Sprintf("%s=%v;", k, v))
	}
	sb.WriteString("}")

	sb.WriteString(fmt.Sprintf("%v", value))
	log.Printf("metrics> %s\n", sb.String())
	return nil
}

type localMetricsClientFactory struct {
}

func (l *localMetricsClientFactory) Create(
	prefix string,
	opt Option,
) (Client, error) {
	return &localMetricsClient{
		prefix:     prefix,
		constLabel: make(map[string]Option),
	}, nil
}

func init() {
	AddClientFactory("local", &localMetricsClientFactory{})
}
