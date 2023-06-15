package util

import (
	"strconv"
)

func ToInt(v interface{}, convertStr bool) (int64, bool) {
	if v == nil {
		return 0, false
	}

	switch val := v.(type) {
	case int:
		return int64(val), true

	case int8:
		return int64(val), true

	case int16:
		return int64(val), true

	case int32:
		return int64(val), true

	case int64:
		return val, true

	case uint:
		return int64(val), true

	case uint8:
		return int64(val), true

	case uint16:
		return int64(val), true

	case uint32:
		return int64(val), true

	case uint64:
		return int64(val), true

	case float32:
		return int64(val), true

	case float64:
		return int64(val), true

	case bool:
		if val {
			return int64(1), true
		} else {
			return int64(0), true
		}

	case string:
		if convertStr {
			i, err := strconv.ParseInt(val, 10, 64)
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
