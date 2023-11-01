package datastar

import (
	"fmt"

	"github.com/delaneyj/toolbelt"
	"github.com/delaneyj/toolbelt/gomps"
)

const (
	GET_ACTION    = "$$get"
	POST_ACTION   = "$$post"
	PUT_ACTION    = "$$put"
	PATCH_ACTION  = "$$patch"
	DELETE_ACTION = "$$delete"
)

func Header(header, expression string) gomps.NODE {
	return gomps.DATA("header-"+toolbelt.Kebab(header), expression)
}

func FetchURL(expression string) gomps.NODE {
	return gomps.DATA("fetch-url", expression)
}

func ServerSentEvents(expression string) gomps.NODE {
	return gomps.DATA("sse", fmt.Sprintf(`'%s'`, expression))
}

func FragmentSelector(querySelector string) gomps.NODE {
	return gomps.DATA("fragment-selector", querySelector)
}

type FragmentSwapType string

const (
	FragmentSwapMorph   FragmentSwapType = "morph"
	FragmentSwapInner   FragmentSwapType = "inner"
	FragmentSwapOuter   FragmentSwapType = "outer"
	FragmentSwapPrepend FragmentSwapType = "prepend"
	FragmentSwapAppend  FragmentSwapType = "append"
	FragmentSwapBefore  FragmentSwapType = "before"
	FragmentSwapAfter   FragmentSwapType = "after"
	FragmentSwapDelete  FragmentSwapType = "delete"
)

func FragmentSwap(swapType FragmentSwapType) gomps.NODE {
	return gomps.DATA("fragment-swap", string(swapType))
}
