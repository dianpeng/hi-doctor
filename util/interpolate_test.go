package util

import (
	"testing"
)

type testInterpPart struct {
	isScript bool
	data     string
}

func TestInterpolationParseEscape(t *testing.T) {
	{
		case1 := "$$"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 1 {
			t.Fatalf("invalid pieces size")
		}
		if list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "$" {
			t.Fatalf("invalid pieces value")
		}
	}
	{
		case1 := "$$<<"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 1 {
			t.Fatalf("invalid pieces size")
		}
		if list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "$<<" {
			t.Fatalf("invalid pieces value")
		}
	}
	{
		case1 := "$$<"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 1 {
			t.Fatalf("invalid pieces size")
		}
		if list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "$<" {
			t.Fatalf("invalid pieces value")
		}
	}
	{
		case1 := ">>"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 1 {
			t.Fatalf("invalid pieces size")
		}
		if list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != ">>" {
			t.Fatalf("invalid pieces value")
		}
	}
	{
		case1 := "$<<>>"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 1 {
			t.Fatalf("invalid pieces size")
		}
		if !list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "" {
			t.Fatalf("invalid pieces value")
		}
	}
}

func TestInterpolationParseMixed(t *testing.T) {
	{
		case1 := "$<<a1>>$<<a2>>"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 2 {
			t.Fatalf("invalid pieces size")
		}
		if !list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "a1" {
			t.Fatalf("invalid pieces value")
		}

		if !list[1].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[1].data != "a2" {
			t.Fatalf("invalid pieces value")
		}
	}

	{
		case1 := "$<<a1>>a$<<a2>>"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 3 {
			t.Fatalf("invalid pieces size")
		}
		if !list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "a1" {
			t.Fatalf("invalid pieces value")
		}

		if list[1].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[1].data != "a" {
			t.Fatalf("invalid pieces value")
		}

		if !list[2].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[2].data != "a2" {
			t.Fatalf("invalid pieces value")
		}
	}

	{
		case1 := "al$<<a1>>a$<<a2>>ar"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 5 {
			t.Fatalf("invalid pieces size")
		}

		if list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "al" {
			t.Fatalf("invalid pieces value")
		}

		if !list[1].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[1].data != "a1" {
			t.Fatalf("invalid pieces value")
		}

		if list[2].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[2].data != "a" {
			t.Fatalf("invalid pieces value")
		}

		if !list[3].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[3].data != "a2" {
			t.Fatalf("invalid pieces value")
		}

		if list[4].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[4].data != "ar" {
			t.Fatalf("invalid pieces value")
		}

	}
}

func TestInterpolationParse1(t *testing.T) {
	// case 1, no interpolation part
	{
		case1 := "abcdef"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 1 {
			t.Fatalf("invalid pieces size")
		}
		if list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "abcdef" {
			t.Fatalf("invalid pieces value")
		}
	}

	// case2.1
	{
		case1 := "abcdef$<<abc>>"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 2 {
			t.Fatalf("invalid pieces size")
		}
		if list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "abcdef" {
			t.Fatalf("invalid pieces value")
		}

		if !list[1].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[1].data != "abc" {
			t.Fatalf("invalid pieces value")
		}
	}

	// case2.2
	{
		case1 := "$<<abc>>abcdef"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 2 {
			t.Fatalf("invalid pieces size")
		}
		if list[1].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[1].data != "abcdef" {
			t.Fatalf("invalid pieces value")
		}

		if !list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "abc" {
			t.Fatalf("invalid pieces value")
		}
	}

	// case2.3
	{
		case1 := "$<<abc>>"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 1 {
			t.Fatalf("invalid pieces size")
		}
		if !list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "abc" {
			t.Fatalf("invalid pieces value")
		}
	}

	// case2.4
	{
		case1 := "xxx1$<<abc>>xxx2"
		list := []testInterpPart{}

		ForeachInterpolation(
			case1,
			func(p string, is bool) error {
				list = append(list, testInterpPart{
					isScript: is,
					data:     p,
				})
				return nil
			},
		)

		if len(list) != 3 {
			t.Fatalf("invalid pieces size")
		}
		if list[0].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[0].data != "xxx1" {
			t.Fatalf("invalid pieces value")
		}

		if !list[1].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[1].data != "abc" {
			t.Fatalf("invalid pieces value")
		}

		if list[2].isScript {
			t.Fatalf("invalid pieces type")
		}
		if list[2].data != "xxx2" {
			t.Fatalf("invalid pieces value")
		}
	}
}
