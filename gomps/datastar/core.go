package datastar

import (
	"strings"

	"github.com/delaneyj/toolbelt/gomps"
	"github.com/goccy/go-json"
)

func MergeStore(m any) gomps.NODE {
	b, err := json.MarshalIndent(m, " ", "")
	if err != nil {
		panic(err)
	}
	s := string(b)
	s = strings.ReplaceAll(s, "\"", "'")

	return gomps.ATTR_RAW("data-merge-store", s)
}

func Ref(name string) gomps.NODE {
	return gomps.DATA("ref", name)
}
