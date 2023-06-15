package storage

import (
	"github.com/dianpeng/hi-doctor/util"
)

type intPrimitive struct {
	value int64
}

func (i *intPrimitive) toInt(x interface{}) (int64, bool) {
	return util.ToInt(x, false)
}

func (i *intPrimitive) Type() string {
	return "int"
}

func (i *intPrimitive) Get() interface{} {
	return i.value
}

func (i *intPrimitive) Set(x interface{}) bool {
	val, ok := i.toInt(x)
	if ok {
		i.value = val
		return true
	} else {
		return false
	}
}

func (i *intPrimitive) CheckedAdd(value interface{}, threshold interface{}) bool {
	if val, ok := i.toInt(value); ok {
		if thr, ok := i.toInt(threshold); ok {
			newV := val + i.value
			i.value = newV
			return newV > thr
		}
	}
	return false
}

func (i *intPrimitive) CheckedSub(value interface{}, threshold interface{}) bool {
	if val, ok := i.toInt(value); ok {
		if thr, ok := i.toInt(threshold); ok {
			newV := val - i.value
			i.value = newV
			return newV < thr
		}
	}
	return false
}

func NewInt(i int64) Primitive {
	return &intPrimitive{
		value: i,
	}
}
