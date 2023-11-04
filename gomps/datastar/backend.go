package datastar

import (
	"fmt"

	"github.com/delaneyj/toolbelt"
	"github.com/delaneyj/toolbelt/gomps"
	"github.com/valyala/bytebufferpool"
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

const FragmentTarget = "self"

func RenderFragment(sse *toolbelt.ServerSentEventsHandler, querySelector string, swap FragmentSwapType, child gomps.NODE) error {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	if err := child.Render(buf); err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}
	sse.Send(
		buf.String(),
		toolbelt.SSEEventId(querySelector),
		toolbelt.SSEEventEvent(string(swap)),
		toolbelt.SSEEventRetry(0),
		toolbelt.SSEEventSkipMinBytesCheck(true),
	)
	return nil
}
