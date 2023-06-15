package storage

import (
	"github.com/dianpeng/hi-doctor/util"
)

const (
	mapTyInt = iota
	mapTyReal
	mapTyBoolean
)

type mapImpl struct {
	ty int
	m  map[string]Primitive
}

func (m *mapImpl) Type() string {
	switch m.ty {
	case mapTyInt:
		return "map[int]"
	case mapTyReal:
		return "map[real]"
	case mapTyBoolean:
		return "map[bool]"
	default:
		panic("unknown type")
		return "unknown"
	}
}

func (m *mapImpl) Get(key string) Primitive {
	x, ok := m.m[key]
	if !ok {
		x = m.init()
		m.m[key] = x
	}

	return x
}

func (m *mapImpl) newPrimitive(x interface{}) Primitive {
	switch m.ty {
	case mapTyInt:
		vv, ok := util.ToInt(x, false)
		if !ok {
			return nil
		}
		return NewInt(vv)
	case mapTyReal:
		vv, ok := util.ToReal(x, false)
		if !ok {
			return nil
		}
		return NewReal(vv)
	case mapTyBoolean:
		vv, ok := x.(bool)
		if !ok {
			return nil
		}
		return NewBoolean(vv)
	default:
		return nil
	}
}

func (m *mapImpl) Set(key string, x interface{}) Primitive {
	val := m.newPrimitive(x)
	m.m[key] = val
	return val
}

func (m *mapImpl) init() Primitive {
	switch m.ty {
	case mapTyInt:
		return NewInt(int64(0))
	case mapTyReal:
		return NewReal(0.0)
	case mapTyBoolean:
		return NewBoolean(false)
	default:
		panic("unknown type")
		return NewInt(int64(0))
	}
}

func NewMapInt() Map {
	return &mapImpl{
		ty: mapTyInt,
		m:  make(map[string]Primitive),
	}
}

func NewMapReal() Map {
	return &mapImpl{
		ty: mapTyReal,
		m:  make(map[string]Primitive),
	}
}

func NewMapBoolean() Map {
	return &mapImpl{
		ty: mapTyBoolean,
		m:  make(map[string]Primitive),
	}
}
