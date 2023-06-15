package storage

// For storage object, we add yet another abstraction here. All the storage
// object internally will just be a interface and this allow us to easily
// work with expression engine. The only downside is that the user has to
// call method to get its actual value out.

type Storage interface {
	Type() string
}

type Primitive interface {
	Type() string
	Get() interface{}
	Set(interface{}) bool
	CheckedAdd(interface{}, interface{}) bool
	CheckedSub(interface{}, interface{}) bool
}

type Map interface {
	Type() string
	Get(string) Primitive
	Set(string, interface{}) Primitive
}

func NewPrimitive(v interface{}) (Primitive, bool) {
	if v == nil {
		return nil, false
	}

	switch val := v.(type) {
	case int:
		return NewInt(int64(val)), true

	case int8:
		return NewInt(int64(val)), true

	case int16:
		return NewInt(int64(val)), true

	case int32:
		return NewInt(int64(val)), true

	case int64:
		return NewInt(val), true

	case uint:
		return NewInt(int64(val)), true

	case uint8:
		return NewInt(int64(val)), true

	case uint16:
		return NewInt(int64(val)), true

	case uint32:
		return NewInt(int64(val)), true

	case uint64:
		return NewInt(int64(val)), true

	case float32:
		return NewReal(float64(val)), true

	case float64:
		return NewReal(val), true

	case bool:
		return NewBoolean(val), true

	default:
		return nil, false
	}
}
