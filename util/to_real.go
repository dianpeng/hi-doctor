package util

import (
	"strconv"
)

func ToReal(v interface{}, convertStr bool) (float64, bool) {
	if v == nil {
		return 0.0, false
	}

	switch val := v.(type) {
	case int:
		return float64(val), true

	case int8:
		return float64(val), true

	case int16:
		return float64(val), true

	case int32:
		return float64(val), true

	case int64:
		return float64(val), true

	case uint:
		return float64(val), true

	case uint8:
		return float64(val), true

	case uint16:
		return float64(val), true

	case uint32:
		return float64(val), true

	case uint64:
		return float64(val), true

	case float32:
		return float64(val), true

	case float64:
		return float64(val), true

	case bool:
		if val {
			return float64(1.0), true
		} else {
			return float64(0.0), true
		}

	case string:
		if convertStr {
			i, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return 0, false
			} else {
				return i, true
			}
		} else {
			return 0, false
		}

	default:
		return 0, false
	}
}
