package exec

import (
	"encoding/json"
	"github.com/dianpeng/hi-doctor/dvar"
	// "fmt"
	"net"
)

type InspectionTargetItem struct {
	Name     string                 `json:"name"` // target name, descriptive if any
	AnyOther map[string]interface{} `json:"-"`    // any other internal information
}

type InspectionTarget []*InspectionTargetItem

func (i *InspectionTargetItem) SetupEnv(env *dvar.EvalEnv) {
	env.Set("target", "name", dvar.NewStringVal(i.Name))
	for k, v := range i.AnyOther {
		env.Set("target", k, dvar.NewInterfaceVal(v))
	}
}

func (i *InspectionTargetItem) DelEnv(env *dvar.EvalEnv) {
	env.Del("target", "name")
	for k, _ := range i.AnyOther {
		env.Del("target", k)
	}
}

func getStrField(input map[string]interface{}, field string) string {
	if v, ok := input[field]; ok {
		if vv, ok := v.(string); ok {
			return vv
		}
	}
	return ""
}

func inspectionTargetJsonV1(input string) InspectionTarget {
	inputByte := []byte(input)
	target := []map[string]interface{}{}
	if err := json.Unmarshal(inputByte, &target); err != nil {
		return nil
	}

	outList := InspectionTarget{}
	for _, t := range target {
		name := getStrField(t, "name")
		other := t

		// A special entry, ie Hostname field. If a hostname field with string
		// content shows up, the inspection target materialization process will try
		// to perform DNS lookup on the specific field
		if maybeHostname, ok := other["hostname"]; ok {
			if hostname, ok := maybeHostname.(string); ok {
				ips, err := net.LookupIP(hostname)
				if err != nil {
					continue
				}

				for _, ip := range ips {
					x := make(map[string]interface{})
					x["ip"] = ip.String()
					for k, v := range other {
						x[k] = v
					}
					outList = append(outList, &InspectionTargetItem{
						Name:     name,
						AnyOther: x,
					})
				}
				continue
			}
		}

		delete(other, "name")

		outList = append(outList, &InspectionTargetItem{
			Name:     name,
			AnyOther: other,
		})
	}

	// Rest cases are just normally handled
	return outList
}

func getInspectionTarget(
	rawInput string,
	format string,
) InspectionTarget {
	switch format {
	default:
		return inspectionTargetJsonV1(rawInput)
	}
}
