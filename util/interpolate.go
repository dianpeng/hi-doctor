package util

import (
	"fmt"
	"strings"
)

const (
	tkText   = 0
	tkStart  = 1
	tkScript = 2
	tkErr    = 3
	tkEof    = 4
)

const eof rune = -1

type interpLexer struct {
	raw      string
	data     []rune
	pos      int
	tk       int
	payload  string
	err      error
	inScript bool
}

func (l *interpLexer) indexAt(value rune, pos int) int {
	// rune comparison here
	for idx := pos; idx < len(l.data); idx++ {
		if l.data[idx] == value {
			return idx
		}
	}
	return -1
}

func (l *interpLexer) diaginfo(pos int) string {
	return fmt.Sprintf("around character position(%d)", pos)
}

func (l *interpLexer) errAt(msg string, pos int) error {
	info := l.diaginfo(pos)
	return fmt.Errorf("%s]: %s", info, msg)
}

func (l *interpLexer) errTk(msg string, pos int) int {
	l.err = l.errAt(msg, pos)
	l.tk = tkErr
	return tkErr
}

func (l *interpLexer) nextText(from int) (string, int, error) {
	idx := l.indexAt('$', from)

	if idx == -1 {
		return string(l.data[from:]), len(l.data), nil
	} else {

		if idx+1 == len(l.data) {
			// there's no pending character after the $, which means this is an error
			return "", -1, l.errAt("unfinished token after '$'", idx)
		} else {
			nchr := l.data[idx+1] // get the next token after the $ which can
			// be escape sequences

			var buf strings.Builder
			buf.WriteString(string(l.data[from:idx]))

			// escape sequences
			switch nchr {
			case '$':
				buf.WriteRune('$')
				// after escape, we need to continue searching until we hit the ${< as
				// script marker start
				data, pos, err := l.nextText(idx + 2)
				if err != nil {
					return "", -1, err
				}

				buf.WriteString(data)
				return buf.String(), pos, nil

			case '<':
				if idx+2 < len(l.data) && l.data[idx+2] == '<' {
					if idx == from {
						// means starts with ${< instead of text, return special marker
						return "", 0, nil
					} else {
						return buf.String(), idx, nil
					}
				} else {
					// ${ is invalid escape characters
					return "", -1, l.errAt("invalid escape token", idx)
				}

			default:
				return "", -1, l.errAt("invalid escape token", idx)
			}
		}
	}
}

func (l *interpLexer) nextScript() int {
	// switch to a character by character based lexer to safely scan cross the
	// scripting part.
	pos := l.pos

	for ; pos < len(l.data); pos++ {
		c := l.data[pos]
		switch c {
		case '\'', '"':
			npos := l.nextScriptStr(pos)
			if npos < 0 {
				return tkErr
			} else {
				pos = npos - 1
			}
			break

		case '>':
			if pos+1 == len(l.data) {
				return l.errTk("unfinished string interpolation part", pos)
			}

			nchr := l.data[pos+1]
			if nchr == '>' {
				// script part done
				l.payload = string(l.data[l.pos:pos])
				l.pos = pos + 2
				l.tk = tkScript
				l.inScript = false
				return tkScript
			} else {
				break
			}

		default:
			break
		}
	}

	return l.errTk("unfinished string interpolation", pos)
}

func (l *interpLexer) nextScriptStr(pos int) int {
	quote := l.data[pos]
	pos++
	ch := l.data[pos]
	for ch != quote {
		if ch == '\n' || ch == eof {
			return l.errTk("unfinished string literal", pos)
		}
		if ch == '\\' {
			if pos+1 == len(l.data) {
				return l.errTk("unfinished string literal, escape part", pos)
			}
			// assume the escape are done correctly, we don't care about it though
			pos += 2
		} else {
			pos++
		}

		if pos == len(l.data) {
			return l.errTk("unfinished string literal", pos)
		}
		ch = l.data[pos]
	}

	return pos + 1
}

func (l *interpLexer) next() int {
	if l.tk == tkErr || l.tk == tkEof {
		return l.tk
	} else if l.pos == len(l.data) {
		l.tk = tkEof
		return tkEof
	}

	if l.inScript {
		return l.nextScript()
	} else {
		data, npos, err := l.nextText(l.pos)
		if err != nil {
			l.err = err
			l.tk = tkErr
			return tkErr
		}

		if npos == 0 {
			// ${< at current position, and we just need to skip it
			l.tk = tkStart
			l.pos += 3
			l.inScript = true
			return tkStart
		}

		l.payload = data
		l.tk = tkText
		l.pos = npos
		return tkText
	}
}

func foreachInterp(
	data string,
	cb func(string, bool) error,
) error {
	l := &interpLexer{
		raw:      data,
		data:     []rune(data),
		pos:      0,
		inScript: false,
	}

LOOP:
	for {
		tk := l.next()
		switch tk {
		case tkText:
			if err := cb(l.payload, false); err != nil {
				return err
			}
			break

		case tkStart:
			break

		case tkScript:
			if err := cb(l.payload, true); err != nil {
				return err
			}
			break

		case tkErr:
			return l.err

		case tkEof:
			break LOOP

		default:
			panic("unreachable")
			break
		}
	}

	return nil
}

func RenderInterpolation(
	payload string,
	eval func(string) (string, error),
) (string, error) {
	buf := &strings.Builder{}

	err := foreachInterp(
		payload,
		func(data string, isScript bool) error {
			if isScript {
				str, err := eval(data)
				if err != nil {
					return err
				}
				buf.WriteString(str)
			} else {
				buf.WriteString(data)
			}
			return nil
		},
	)

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ForeachInterpolation(
	data string,
	cb func(string, bool) error,
) error {
	return foreachInterp(data, cb)
}
