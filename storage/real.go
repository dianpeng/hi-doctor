package storage

import (
	"github.com/dianpeng/hi-doctor/util"
)

type realPrimitive struct {
	value float64
}

func (i *realPrimitive) toReal(x interface{}) (float64, bool) {
	return util.ToReal(x, false)
}

func (i *realPrimitive) Type() string {
	return "real"
}

func (i *realPrimitive) Get() interface{} {
	return i.value
}

func (i *realPrimitive) Set(x interface{}) bool {
	val, ok := i.toReal(x)
	if ok {
		i.value = val
		return true
	} else {
		return false
	}
}

func (i *realPrimitive) CheckedAdd(value interface{}, threshold interface{}) bool {
	if val, ok := i.toReal(value); ok {
		if thr, ok := i.toReal(threshold); ok {
			newV := val + i.value
			i.value = newV
			return newV > thr
		}
	}
	return false
}

func (i *realPrimitive) CheckedSub(value interface{}, threshold interface{}) bool {
	if val, ok := i.toReal(value); ok {
		if thr, ok := i.toReal(threshold); ok {
			newV := val - i.value
			i.value = newV
			return newV < thr
		}
	}
	return false
}

func NewReal(v float64) Primitive {
	return &realPrimitive{
		value: v,
	}
}
