package dvar

import (
	"fmt"
	"strings"

	"github.com/dianpeng/hi-doctor/util"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

// A DVar is a specialized Var which requires evaluation when you want to get
// its output or value. It requires a EvalEnv which serves as a blackboard
// to get/share states.
//
// Essentially DVar is just a thin wrapper around expression engine plus some
// interpolation features. The reason is inside of yaml file, user cannot
// specify which part is *code/expression* which requiers evaluation and which
// part is just plain string. The syntax simple does not allow so. In order to
// resolve this issue, we introduce a thin syntax around the native string.
//
// $:expression -> this is an expression, ie, anythin inside of it is expression
// otherwise    -> a string or string interpolation. Which will be assembled
//              -> to valid expression. Finally everything will just become an
//                 expression and compiled down to bytecode for evaluation
//
// Notes, in our system, all the expression evaluation output will be Val not
// interface{}/any.

type DVar struct {
	prog *vm.Program // the underlying program, if it is null then it is literal
	lit  string      // the literal behind the scene
}

const (
	StringContext = iota
	ScriptContext
)

type CodeBlock []DVar

func CompileCodeBlock(context string, input []string) (CodeBlock, error) {
	out := CodeBlock{}
	for i, v := range input {
		if dv, err := NewDVarScriptContext(v); err != nil {
			return nil, fmt.Errorf("%s[%d] compile fail: %s", context, i, err)
		} else {
			out = append(out, dv)
		}
	}
	return out, nil
}

func (d *DVar) IsLiteral() bool {
	return d.prog == nil
}

func (d *DVar) Value(env *EvalEnv) (Val, error) {
	if d.IsLiteral() {
		return NewStringVal(d.lit), nil
	} else {
		output, err := expr.Run(d.prog, env.ExprEnv())
		if err != nil {
			return Val{}, fmt.Errorf("evaluation of script{ %s } failed: %s", d.lit, err)
		} else {
			vv := NewInterfaceVal(output)
			return vv, nil
		}
	}
}

type strpiece struct {
	data     string
	isScript bool
}

func NewDVarLit(lit string) DVar {
	return DVar{
		prog: nil,
		lit:  lit,
	}
}

func NewDVarScriptContext(data string) (DVar, error) {
	return NewDVar(data, ScriptContext)
}

func NewDVarStringContext(data string) (DVar, error) {
	return NewDVar(data, StringContext)
}

func NewDVar(data string, context int) (DVar, error) {
	// 1) if the data is empty, just stores it as literal
	if data == "" {
		// shortcut for empty string
		return DVar{
			lit: "",
		}, nil
	}

	// 2) code indication, try to compile it as an expression snippet
	if len(data) >= 3 && data[0] == '$' && data[1] == '{' && data[len(data)-1] == '}' {
		// shortcut for expression
		p, err := expr.Compile(data[2 : len(data)-1])
		if err != nil {
			return DVar{}, err
		}
		return DVar{
			prog: p,
			lit:  data,
		}, nil
	}

	// 3) string literal, ie just directly compile it as string literal
	if len(data) >= 3 && data[0] == '$' && data[1] == '(' && data[len(data)-1] == ')' {
		return DVar{
			lit: data[2 : len(data)-1],
		}, nil
	}

	// 4) based on evaluation context to decide what to do next
	switch context {
	case StringContext:
		return newDVarFromStringInterp(data)

	case ScriptContext:
		// just wrap everything as script and compile it
		p, err := expr.Compile(data)
		if err != nil {
			return DVar{}, err
		}
		return DVar{
			prog: p,
			lit:  data,
		}, nil

	default:
		unreachable("invalid context")
		return DVar{}, nil
	}
}

func newDVarFromStringInterp(data string) (DVar, error) {
	pieceList := []strpiece{}
	hasScript := false
	hasScriptPtr := &hasScript

	// string or string interpolation
	util.ForeachInterpolation(
		data,
		func(d string, is bool) error {
			pieceList = append(pieceList, strpiece{
				data:     d,
				isScript: is,
			})
			if is {
				*hasScriptPtr = true
			}
			return nil
		},
	)

	if hasScript {
		piece := []string{}
		// assemble everything as string concatenation, ie the program will become
		// str#0 + str#1 ... str#N
		for _, v := range pieceList {
			if v.isScript {
				piece = append(piece, fmt.Sprintf("(%s)", v.data))
			} else {
				vv, ok := util.Unescape(v.data, '\'')
				must(ok, "must be valid sequences")
				piece = append(piece, fmt.Sprintf("'%s'", vv))
			}
		}
		code := strings.Join(piece, " + ")
		prog, err := expr.Compile(code)
		if err != nil {
			return DVar{}, fmt.Errorf("string interpolation code(%s) compiles failed: %s",
				code, err)
		}
		return DVar{
			prog: prog,
			lit:  code,
		}, nil
	} else {
		b := new(strings.Builder)
		for _, v := range pieceList {
			must(!v.isScript, "should never be script")
			b.WriteString(v.data)
		}

		return DVar{
			prog: nil,
			lit:  b.String(),
		}, nil
	}
}
