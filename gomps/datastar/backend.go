package datastar

import (
	"fmt"
	"time"

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

func FetchURLF(format string, args ...interface{}) gomps.NODE {
	return FetchURL(fmt.Sprintf(format, args...))
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
	FragmentSelectorSelf      = "self"
	FragmentSelectorUseID     = ""
	FragmentEventTypeFragment = "datastar-fragment"
)

type RenderFragmentOptions struct {
	QuerySelector  string
	Merge          FragmentMergeType
	SettleDuration time.Duration
}
type RenderFragmentOption func(*RenderFragmentOptions)

func WithQuerySelector(selector string) RenderFragmentOption {
	return func(o *RenderFragmentOptions) {
		o.QuerySelector = selector
	}
}

func WithMergeType(merge FragmentMergeType) RenderFragmentOption {
	return func(o *RenderFragmentOptions) {
		o.Merge = merge
	}
}

func WithSettleDuration(d time.Duration) RenderFragmentOption {
	return func(o *RenderFragmentOptions) {
		o.SettleDuration = d
	}
}

func RenderFragment(sse *toolbelt.ServerSentEventsHandler, child gomps.NODE, opts ...RenderFragmentOption) error {
	options := &RenderFragmentOptions{
		QuerySelector:  FragmentSelectorSelf,
		Merge:          FragmentMergeMorphElement,
		SettleDuration: 0,
	}
	for _, opt := range opts {
		opt(options)
	}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	if err := child.Render(buf); err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}

	dataRows := []string{
		fmt.Sprintf("selector %s", options.QuerySelector),
		fmt.Sprintf("merge %s", options.Merge),
		fmt.Sprintf("settle %d", options.SettleDuration.Milliseconds()),
		fmt.Sprintf("html %s", buf.String()),
	}

	sse.SendMultiData(
		dataRows,
		toolbelt.WithSSEEvent(FragmentEventTypeFragment),
		toolbelt.WithSSERetry(0),
		toolbelt.WithSSESkipMinBytesCheck(true),
	)
	return nil
}
