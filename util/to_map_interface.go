package util

import (
	"encoding/json"
)

func ToMapInterface(x interface{}) map[string]interface{} {
	out, _ := json.Marshal(x)
	ret := make(map[string]interface{})
	json.Unmarshal(out, &ret)
	return ret
}
