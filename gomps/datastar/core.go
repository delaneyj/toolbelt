package datastar

import (
	"github.com/delaneyj/toolbelt/gomps"
	"github.com/goccy/go-json"
)

func MergeStore(m any) gomps.NODE {
	b, err := json.MarshalIndent(m, " ", "")
	if err != nil {
		panic(err)
	}
	s := string(b)

	return gomps.DATA("merge-store", s)
}

func Ref(name string) gomps.NODE {
	return gomps.DATA("ref", name)
}
