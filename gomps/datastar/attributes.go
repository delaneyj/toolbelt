package datastar

import (
	"fmt"

	"github.com/delaneyj/toolbelt/gomps"
)

func Model(expression string) gomps.NODE {
	return gomps.DATA("model", expression)
}

func Text(expression string) gomps.NODE {
	return gomps.DATA("text", expression)
}

func TextF(format string, args ...interface{}) gomps.NODE {
	return gomps.DATA("text", fmt.Sprintf(format, args...))
}

func On(eventName, expression string) gomps.NODE {
	return gomps.DATA("on-"+eventName, expression)
}

func Focus(expression string) gomps.NODE {
	return gomps.DATA("focus")
}
