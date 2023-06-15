package util

import (
	"strings"
	"unicode/utf8"
)

func Unescape(value string, quote rune) (string, bool) {
	sb := new(strings.Builder)
	cursor := 0
	size := len(value)
	for cursor < size {
		r, size := utf8.DecodeRuneInString(value[cursor:])
		if !utf8.ValidRune(r) {
			return "", false
		} else {
			switch r {
			case '\a':
				sb.WriteString("\\a")
				break
			case '\b':
				sb.WriteString("\\b")
				break
			case '\v':
				sb.WriteString("\\v")
				break
			case '\f':
				sb.WriteString("\\f")
				break
			case '\n':
				sb.WriteString("\\n")
				break
			case '\r':
				sb.WriteString("\\r")
				break
			case '\t':
				sb.WriteString("\\t")
				break
			case quote:
				sb.WriteRune('\\')
				sb.WriteRune(quote)
				break
				break
			case '\\':
				sb.WriteString("\\\\")
				break
			default:
				sb.WriteRune(r)
				break
			}
		}

		cursor += size
	}
	return sb.String(), true
}
