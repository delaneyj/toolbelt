package toolbelt

import (
	"strings"

	"github.com/iancoleman/strcase"
)

func Pascal(s string) string {
	return strcase.ToCamel(s)
}

func Camel(s string) string {
	return strcase.ToLowerCamel(s)
}

func Snake(s string) string {
	return strcase.ToSnake(s)
}

func ScreamingSnake(s string) string {
	return strcase.ToScreamingSnake(s)
}

func Kebab(s string) string {
	return strcase.ToKebab(s)
}

func Upper(s string) string {
	return strings.ToUpper(s)
}

func Lower(s string) string {
	return strings.ToLower(s)
}

type CasedFn func(string) string

func Cased(s string, fn ...CasedFn) string {
	for _, f := range fn {
		s = f(s)
	}
	return s
}

type CasedString struct {
	Original string
	Pascal   string
	Camel    string
	Snake    string
	Kebab    string
	Upper    string
	Lower    string
}

func ToCasedString(s string) CasedString {
	return CasedString{
		Original: s,
		Pascal:   Pascal(s),
		Camel:    Camel(s),
		Snake:    Snake(s),
		Kebab:    Kebab(s),
		Upper:    Upper(s),
		Lower:    Lower(s),
	}
}
