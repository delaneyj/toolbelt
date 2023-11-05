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

type FragmentMergeType string

const (
	FragmentMergeMorphElement     FragmentMergeType = "morph_element"
	FragmentMergeInnerElement     FragmentMergeType = "inner_element"
	FragmentMergeOuterElement     FragmentMergeType = "outer_element"
	FragmentMergePrependElement   FragmentMergeType = "prepend_element"
	FragmentMergeAppendElement    FragmentMergeType = "append_element"
	FragmentMergeBeforeElement    FragmentMergeType = "before_element"
	FragmentMergeAfterElement     FragmentMergeType = "after_element"
	FragmentMergeDeleteElement    FragmentMergeType = "delete_element"
	FragmentMergeUpsertAttributes FragmentMergeType = "upsert_attributes"
)

const (
	FragmentSelectorSelf  = "self"
	FragmentSelectorUseID = ""
)

func RenderFragment(sse *toolbelt.ServerSentEventsHandler, querySelector string, swap FragmentMergeType, child gomps.NODE) error {
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
