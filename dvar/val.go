// A simpler tagged union, just make the type explicit without using reflection
// of go

package dvar

import (
	"fmt"
	"github.com/dianpeng/hi-doctor/util"
	"strconv"
)

const (
	ValNull = iota
	ValInt
	ValReal
	ValString
	ValBoolean

	// go types, stored as opaque interface{}, and we do not know what it is
	ValAny
)

type Val struct {
	ty  int
	val interface{}
}

func (v *Val) Type() int {
	return v.ty
}

func (v *Val) IsInt() bool {
	return v.ty == ValInt
}

func (v *Val) IsReal() bool {
	return v.ty == ValReal
}

func (v *Val) IsString() bool {
	return v.ty == ValString
}

func (v *Val) IsBoolean() bool {
	return v.ty == ValBoolean
}

func (v *Val) IsNull() bool {
	return v.ty == ValNull
}

func (v *Val) IsAny() bool {
	return v.ty == ValAny
}

func (v *Val) GetInt() int64 {
	must(v.ty == ValInt, "must be int")
	vv, _ := v.val.(int64)
	return vv
}

func (v *Val) GetReal() float64 {
	must(v.ty == ValReal, "must be real")
	vv, _ := v.val.(float64)
	return vv
}

func (v *Val) GetString() string {
	must(v.ty == ValString, "must be string")
	vv, _ := v.val.(string)
	return vv
}

func (v *Val) GetBoolean() bool {
	must(v.ty == ValBoolean, "must be boolean")
	vv, _ := v.val.(bool)
	return vv
}

func (v *Val) GetAny() interface{} {
	must(v.ty == ValAny, "must be any")
	return v.val
}

func (v *Val) SetInt(i int64) {
	v.ty = ValInt
	v.val = i
}

func (v *Val) SetReal(r float64) {
	v.ty = ValReal
	v.val = r
}

func (v *Val) SetString(r string) {
	v.ty = ValString
	v.val = r
}

func (v *Val) SetBoolean(b bool) {
	v.ty = ValBoolean
	v.val = b
}

func (v *Val) SetNull() {
	v.ty = ValNull
	v.val = nil
}

func (v *Val) SetAny(x interface{}) {
	v.ty = ValAny
	v.val = x
}

func (val *Val) Int() (int64, bool) {
	v := val.val
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
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, false
		} else {
			return i, true
		}

	default:
		return 0, false
	}
}

func (val *Val) Real() (float64, bool) {
	v := val.val
	if v == nil {
		return 0, false
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
		i, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, false
		} else {
			return i, true
		}

	default:
		return 0, false
	}
}

func (v *Val) String() string {
	switch v.ty {
	case ValInt:
		return fmt.Sprintf("%d", v.GetInt())

	case ValReal:
		return fmt.Sprintf("%f", v.GetReal())

	case ValString:
		return v.GetString()

	case ValBoolean:
		return fmt.Sprintf("%t", v.GetBoolean())

	case ValNull:
		return "null"

	case ValAny:
		return "any"

	default:
		unreachable("invalid type")
		return ""
	}
}

func (v *Val) Boolean() bool {
	switch v.ty {
	case ValInt:
		return v.GetInt() != 0

	case ValReal:
		return v.GetReal() != 0.0

	case ValBoolean:
		return v.GetBoolean()

	case ValNull:
		return false

	case ValString:
		return true // NOTES(dpeng): For string, always return *true*

	case ValAny:
		return true

	default:
		unreachable("invalid type")
		return false
	}
}

func (v *Val) Debug() string {
	switch v.ty {
	case ValInt:
		return fmt.Sprintf("%d", v.GetInt())

	case ValReal:
		return fmt.Sprintf("%f", v.GetReal())

	case ValString:
		return v.GetString()

	case ValBoolean:
		return fmt.Sprintf("%t", v.GetBoolean())

	case ValNull:
		return "null"

	case ValAny:
		return fmt.Sprintf("%v", v.val)

	default:
		unreachable("invalid type")
		return ""
	}
}

func (v *Val) Interface() interface{} {
	return v.val
}

// Convinient methods
func (v *Val) Port() (uint16, bool) {
	intV, ok := v.Int()
	if !ok {
		return 0, false
	}
	if intV < 0 || intV > 65535 {
		return 0, false
	}
	return uint16(intV), true
}

func (v *Val) PortList() ([]uint16, bool) {
	if v.IsAny() {
		if vv, ok := v.GetAny().([]interface{}); ok {
			out := []uint16{}
			for _, maybePort := range vv {
				if num, ok := util.ToInt(maybePort, false); ok {
					if num >= 0 && num < 65536 {
						out = append(out, uint16(num))
					}
				}
			}
			return out, true
		}
	}
	return nil, false
}

func NewIntVal(i int64) Val {
	return Val{
		ty:  ValInt,
		val: i,
	}
}

func NewRealVal(i float64) Val {
	return Val{
		ty:  ValReal,
		val: i,
	}
}

func NewNullVal() Val {
	return Val{}
}

func NewStringVal(str string) Val {
	return Val{
		ty:  ValString,
		val: str,
	}
}

func NewBooleanVal(b bool) Val {
	return Val{
		ty:  ValBoolean,
		val: b,
	}
}

func NewTrueVal() Val {
	return NewBooleanVal(true)
}

func NewFalseVal() Val {
	return NewBooleanVal(false)
}

func NewAnyVal(x interface{}) Val {
	return Val{
		ty:  ValAny,
		val: x,
	}
}

func NewInterfaceVal(v interface{}) Val {
	if v == nil {
		return NewNullVal()
	}

	switch val := v.(type) {
	case Val:
		return val
	case *Val:
		return *val
	case int:
		return NewIntVal(int64(val))

	case int8:
		return NewIntVal(int64(val))

	case int16:
		return NewIntVal(int64(val))

	case int32:
		return NewIntVal(int64(val))

	case int64:
		return NewIntVal(val)

	case uint:
		return NewIntVal(int64(val))

	case uint8:
		return NewIntVal(int64(val))

	case uint16:
		return NewIntVal(int64(val))

	case uint32:
		return NewIntVal(int64(val))

	case uint64:
		return NewIntVal(int64(val))

	case float32:
		return NewRealVal(float64(val))

	case float64:
		return NewRealVal(val)

	case bool:
		return NewBooleanVal(val)

	case string:
		return NewStringVal(val)

	default:
		return NewAnyVal(v)
	}
}

type ValMap map[string]Val

func NewValMap() ValMap {
	return make(ValMap)
}

func PopulateAssetsMap(x map[string]interface{}) (ValMap, error) {
	env := NewEvalEnv()

	out := make(ValMap)
	for k, v := range x {
		vv := NewInterfaceVal(v)
		if vv.IsString() {
			script := vv.String()
			if dv, err := NewDVarStringContext(script); err != nil {
				return nil, fmt.Errorf("Assets[%s] compilation error %s", k, err)
			} else {
				if val, err := dv.Value(env); err != nil {
					return nil, fmt.Errorf("Assets[%s] evaluation error %s", k, err)
				} else {
					out[k] = val
				}
			}
		} else {
			out[k] = vv
		}
	}

	return out, nil
}
