package storage

type booleanPrimitive struct {
	v bool
}

func (b *booleanPrimitive) Type() string {
	return "boolean"
}

func (b *booleanPrimitive) Get() interface{} {
	return b.v
}

func (b *booleanPrimitive) Set(x interface{}) bool {
	vv, ok := x.(bool)
	if ok {
		b.v = vv
		return true
	}
	return false
}

func (b *booleanPrimitive) CheckedAdd(interface{}, interface{}) bool {
	return false
}

func (b *booleanPrimitive) CheckedSub(interface{}, interface{}) bool {
	return false
}

func NewBoolean(b bool) Primitive {
	return &booleanPrimitive{
		v: b,
	}
}
