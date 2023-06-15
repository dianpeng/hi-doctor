package dvar

import (
	"fmt"

	// for basic internal library support
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"net/http"

	"github.com/dianpeng/hi-doctor/util"
)

type EvalEnv struct {
	data map[string]interface{} // for expr library
}

type fieldMap map[string]interface{}

func (e *EvalEnv) ExprEnv() map[string]interface{} {
	return e.data
}

/* ---------------------------------------------------------------------------
 * Inheritance APIs
 * -------------------------------------------------------------------------*/
func (e *EvalEnv) Inherit(base ValMap) {
	for k, v := range base {
		_, ok := e.data[k]
		if !ok {
			e.data[k] = v.Interface()
		}
	}
}

func (e *EvalEnv) InheritInNamespace(ns string, base ValMap) {
	field := e.getOrCreateField(ns)
	for k, v := range base {
		_, ok := field[k]
		if !ok {
			field[k] = v.Interface()
		}
	}
}

func (e *EvalEnv) InheritFromEnv(base *EvalEnv) {
	for k, v := range base.data {
		_, ok := e.data[k]
		if !ok {
			e.data[k] = v
		}
	}
}

func (e *EvalEnv) getField(field string) fieldMap {
	v, ok := e.data[field]
	if !ok {
		return nil
	}
	vv, ok := v.(fieldMap)
	if !ok {
		return nil
	}
	return vv
}

func (e *EvalEnv) GetNamespace(field string) map[string]interface{} {
	return e.getOrCreateField(field)
}

// Special function used by multiple tasks to achieve store the history and
// the current and the last.
// What the following function does is as following :
//
// 1) Call it as RecordHistoricalResult("namespace", xxx: map[string]interface{})
//
// The function will try
//   1) Append the xxx into a history field under namespace specified
//   2) Set the last field under namespace specified to point to xxx
//   3) inline all key value inside of xxx under namespace specified
//      *name collision should be handled by the caller*

func (e *EvalEnv) RecordHistoricalResult(
	field string,
	stat map[string]interface{},
) {
	ns := e.GetNamespace(field)
	if old, ok := ns["history"]; ok {
		oldV, ok := old.([]map[string]interface{})
		if !ok {
			panic("invalid runtime history interface")
		}
		oldV = append(oldV, stat)
		ns["history"] = oldV
	} else {
		ns["history"] = []map[string]interface{}{stat}
	}
	ns["last"] = stat
	for k, v := range stat {
		ns[k] = v
	}
}

func (e *EvalEnv) getOrCreateField(field string) fieldMap {
	v := e.getField(field)
	if v == nil {
		v = make(fieldMap)
		e.data[field] = v
	}
	return v
}

func (e *EvalEnv) Set(field, key string, v Val) {
	f := e.getOrCreateField(field)
	f[key] = v.Interface()
}

func (e *EvalEnv) GetOrNull(field, key string) Val {
	f := e.getField(field)
	if f == nil {
		return NewNullVal()
	}

	v, ok := f[key]
	if !ok {
		return NewNullVal()
	}
	return NewInterfaceVal(v)
}

func (e *EvalEnv) Get(field, key string) (Val, bool) {
	f := e.getField(field)
	if f == nil {
		return Val{}, false
	}
	v, ok := f[key]
	if !ok {
		return Val{}, false
	}
	return NewInterfaceVal(v), true
}

func (e *EvalEnv) GetOrCrash(field, key string) Val {
	v, ok := e.Get(field, key)
	must(ok, fmt.Sprintf("GetOrCrash(%s, %s)", field, key))
	return v
}

func (e *EvalEnv) GetDef(field, key string, def Val) Val {
	v, ok := e.Get(field, key)
	if !ok {
		return def
	} else {
		return v
	}
}

func (e *EvalEnv) Del(field, key string) {
	f := e.getField(field)
	if f != nil {
		delete(f, key)
	}
}

func NewEvalEnv() *EvalEnv {
	x := &EvalEnv{
		data: make(map[string]interface{}),
	}
	addLibrary(x)
	return x
}

func NewEvalEnvFromBase(base *EvalEnv) *EvalEnv {
	x := NewEvalEnv()
	x.InheritFromEnv(base)
	return x
}

/* ---------------------------------------------------------------------------
 * Evaluation Library Common Library
 * --------------------------------------------------------------------------*/

func addLibrary(env *EvalEnv) {
	addBaseLibraryMisc(env)
	addBaseLibraryOS(env)
	addBaseLibraryPrint(env)
	addBaseLibraryTime(env)
	addBaseLibraryString(env)
	addBaseLibraryRandom(env)
	addBaseLibraryHttp(env)
}

func addBaseLibraryMisc(env *EvalEnv) {
	rawMap := env.ExprEnv()
	rawMap["Str"] = func(xx interface{}) string {
		return fmt.Sprintf("%v", xx)
	}
	rawMap["PrettyStr"] = func(xx interface{}) string {
		b, err := json.MarshalIndent(xx, "", "  ")
		if err != nil {
			return "<N/A>"
		} else {
			return string(b)
		}
	}
}

func addBaseLibraryOS(env *EvalEnv) {
	lib := env.GetNamespace("os")
	lib["Hostname"] = func() interface{} {
		h, err := os.Hostname()
		if err != nil {
			return nil
		}
		return h
	}
	lib["Pwd"] = func() interface{} {
		path, err := os.Getwd()
		if err != nil {
			return nil
		}
		return path
	}
	lib["Executable"] = func() interface{} {
		v, err := os.Executable()
		if err != nil {
			return nil
		}
		return v
	}
	lib["Env"] = os.Getenv
}

func addBaseLibraryPrint(env *EvalEnv) {
	libPrint := env.GetNamespace("print")

	{
		libPrint["Print"] = fmt.Print
		libPrint["Println"] = fmt.Println
		libPrint["Printf"] = fmt.Printf
	}
}

func addBaseLibraryTime(env *EvalEnv) {
	lib := env.GetNamespace("time")
	{
		lib["Now"] = func() int64 {
			return time.Now().Unix()
		}
		lib["NowMicro"] = func() int64 {
			return time.Now().UnixMicro()
		}
		lib["NowMilli"] = func() int64 {
			return time.Now().UnixMilli()
		}
		lib["NowNano"] = func() int64 {
			return time.Now().UnixNano()
		}
	}
}

func addBaseLibraryRandom(env *EvalEnv) {
	lib := env.GetNamespace("rand")
	{
		lib["GenStr"] = func(sz int) string {
			return util.RndStr(sz)
		}
		lib["Gen1K"] = func() string {
			return util.RndStr(1024)
		}
		lib["Gen4K"] = func() string {
			return util.RndStr(1024 * 4)
		}
		lib["Gen16K"] = func() string {
			return util.RndStr(1024 * 16)
		}
		lib["Gen32K"] = func() string {
			return util.RndStr(1024 * 32)
		}
		lib["Gen64K"] = func() string {
			return util.RndStr(1024 * 64)
		}
		lib["Gen128K"] = func() string {
			return util.RndStr(1024 * 128)
		}
		lib["Gen256K"] = func() string {
			return util.RndStr(1024 * 256)
		}
		lib["Gen512K"] = func() string {
			return util.RndStr(1024 * 512)
		}
		lib["Gen1MB"] = func() string {
			return util.RndStr(1024 * 1024)
		}
		lib["Int"] = func() int {
			return rand.Int()
		}
		lib["Real"] = func() float64 {
			return rand.Float64()
		}
	}
}

func addBaseLibraryHttp(env *EvalEnv) {
	libHttp := env.GetNamespace("http")
	{
		libHttp["HeaderGet"] = func(hdr http.Header, key string) string {
			return hdr.Get(key)
		}
		libHttp["HeaderHas"] = func(hdr http.Header, key string) bool {
			return hdr.Get(key) != ""
		}
		libHttp["HeaderDump"] = func(hdr http.Header) string {
			sb := new(strings.Builder)
			for k, v := range hdr {
				sb.WriteString(fmt.Sprintf("%s => %s\n", k, strings.Join(v, ", ")))
			}
			return sb.String()
		}
		libHttp["Get"] = func(
			url string,
			expectStatus int,
		) map[string]interface{} {
			resp, err := http.Get(url)
			out := make(map[string]interface{})

			if err != nil {
				out["resp_err"] = fmt.Sprintf("%s", err)
				out["resp_ok"] = false
				return out
			} else {
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				respErr := ""
				bodyStr := ""
				respOk := false

				if err != nil {
					respOk = false
					respErr = fmt.Sprintf("%s", err)
				} else {
					respOk = resp.StatusCode == expectStatus
					bodyStr = string(body)
				}
				out["resp_ok"] = respOk
				out["resp_header"] = resp.Header
				out["resp_status"] = resp.StatusCode
				out["resp_body"] = bodyStr
				out["resp_err"] = respErr
				return out
			}
		}
	}
}

func addBaseLibraryString(env *EvalEnv) {
	libStr := env.GetNamespace("string")
	{
		libStr["Compare"] = strings.Compare
		libStr["Contains"] = strings.Contains
		libStr["ContainsAny"] = strings.ContainsAny
		libStr["Count"] = strings.Count
		libStr["HasPrefix"] = strings.HasPrefix
		libStr["HasSuffix"] = strings.HasSuffix
		libStr["Trim"] = strings.Trim
		libStr["TrimLeft"] = strings.TrimLeft
		libStr["TrimRight"] = strings.TrimRight
		libStr["TrimSpace"] = strings.TrimSpace
		libStr["TrimSuffix"] = strings.TrimSuffix
		libStr["Length"] = func(x string) int {
			return len(x)
		}
		libStr["Empty"] = func(x string) bool {
			return len(x) == 0
		}
		libStr["Sprintf"] = fmt.Sprintf
	}
}
