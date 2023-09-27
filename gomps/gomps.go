package webui

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/delaneyj/toolbelt"
	json "github.com/goccy/go-json"
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	h "github.com/maragudk/gomponents/html"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	A           = h.A
	ATTR        = g.Attr
	DIV         = h.Div
	BUTTON      = h.Button
	SPAN        = h.Span
	CLS         = h.Class
	META        = h.Meta
	NAME        = h.Name
	CONTENT     = h.Content
	LINK        = h.Link
	REL         = h.Rel
	CHARSET     = h.Charset
	TITLE       = h.TitleEl
	HEAD        = h.Head
	HTML        = h.HTML
	HTML5       = c.HTML5
	DOCTYPE     = h.Doctype
	BODY        = h.Body
	LANG        = h.Lang
	HR          = h.Hr
	HREF        = h.Href
	IMG         = h.Img
	SRC         = h.Src
	SCRIPT      = h.Script
	LABEL       = h.Label
	PLACEHOLDER = h.Placeholder
	TEXTAREA    = h.Textarea
	SELECT      = h.Select
	OPTION      = h.Option
	TABLE       = h.Table
	CAPTION     = h.Caption
	THEAD       = h.THead
	TBODY       = h.TBody
	TR          = h.Tr
	TH          = h.Th
	TD          = h.Td
	STYLE       = h.StyleAttr

	REQUIRED = h.Required()
	DISABLED = h.Disabled()
	SELECTED = h.Selected()
	CHECKED  = g.Attr("checked")
	DEFER    = h.Defer()
	RAW      = g.Raw

	H1 = h.H1
	H2 = h.H2
	H3 = h.H3
	H4 = h.H4
	H5 = h.H5
	H6 = h.H6
	P  = h.P

	PRE  = h.Pre
	CODE = h.Code

	UL = h.Ul
	LI = h.Li

	ID  = h.ID
	ALT = h.Alt

	DETAILS = h.Details
	OPEN    = ATTR("open")
	SUMMARY = h.Summary

	FORM   = h.FormEl
	INPUT  = h.Input
	FOR    = h.For
	TYPE   = h.Type
	VALUE  = h.Value
	ACTION = h.Action
	METHOD = h.Method
	MIN    = h.Min
	MAX    = h.Max
)

type (
	CLSS       = c.Classes
	NODE       = g.Node
	NODES      = []g.Node
	HTML5Props = c.HTML5Props
)
type Highlight struct {
	Language string
	Code     string
	Style    string
	Children NODES
}

var (
	HighlightDefaultStyle = "gruvbox"
)

func (h Highlight) Render(w io.Writer) error {
	lexer := lexers.Get(h.Language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get(HighlightDefaultStyle)

	iter, err := lexer.Tokenise(nil, h.Code)
	if err != nil {
		return err
	}

	formatter := html.New(
		html.TabWidth(2),
		html.WithLineNumbers(true),
		html.Standalone(false),
	)

	buf := &strings.Builder{}
	if err = formatter.Format(buf, style, iter); err != nil {
		return fmt.Errorf("formatting code: %w", err)
	}

	return DIV(
		RAW(buf.String()),
		GRP(h.Children...),
	).Render(w)
}

func HIGHLIGHT(language, code string, children ...NODE) NODE {
	h := Highlight{
		Language: language,
		Code:     code,
		Style:    HighlightDefaultStyle,
	}
	h.Language = language
	h.Code = code
	return h
}

func STYLEF(format string, args ...interface{}) NODE {
	return STYLE(fmt.Sprintf(format, args...))
}

func DATA(name, value string) NODE {
	return ATTR("data-"+name, value)
}

func DATAF(name, format string, args ...interface{}) NODE {
	return DATA(name, fmt.Sprintf(format, args...))
}

func GRP(children ...NODE) NODE {
	return g.Group(children)
}

func CLSF(format string, args ...interface{}) NODE {
	return CLS(fmt.Sprintf(format, args...))
}

func STEP(step int) NODE {
	return ATTR("step", strconv.Itoa(step))
}

func MINI(min int) NODE {
	return ATTR("min", strconv.Itoa(min))
}

func MAXI(max int) NODE {
	return ATTR("max", strconv.Itoa(max))
}

var INTEGERTYPE = GRP(
	TYPE("number"),
	STEP(1),
)

func COLSPAN(colspan int) NODE {
	return ATTR("colspan", strconv.Itoa(colspan))
}

func ViewTransitionName(id string) NODE {
	return h.StyleAttr("view-transition-name: " + id)
}

func IDVT(id string) NODE {
	return GRP(
		ID(id),
		ViewTransitionName(toolbelt.Kebab(id)),
	)
}

func VALUEI[T uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64](v T) NODE {
	return VALUE(strconv.Itoa(int(v)))
}

func TXT(text string) NODE {
	return g.Text(text)
}

func TXTI[T uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64](v T) NODE {
	return TXT(strconv.Itoa(int(v)))
}

func TXTF(format string, args ...interface{}) NODE {
	return TXT(fmt.Sprintf(format, args...))
}

func SAFE(text string) NODE {
	return g.Raw(text)
}

type NodeFn func(children ...NODE) NODE

func TERN(cond bool, ifTrue, ifFalse NodeFn) NODE {
	if cond {
		return ifTrue()
	}
	return ifFalse()
}

func TERNCached(cond bool, ifTrue, ifFalse NODE) NODE {
	if cond {
		return ifTrue
	}
	return ifFalse
}

func EMPTY[T any](arr []T) bool {
	return len(arr) == 0
}

func PREJSON[T any](v T) NODE {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return TXT(err.Error())
	}
	return PRE(SAFE(string(b)))
}

func PREPBJSON(m protoreflect.ProtoMessage) NODE {
	b, err := protojson.Marshal(m)
	if err != nil {
		return TXT(err.Error())
	}
	return PRE(SAFE(string(b)))
}

func RANGE[T any](ts []T, cb func(item T) NODE) NODE {
	var nodes []NODE
	for _, t := range ts {
		nodes = append(nodes, cb(t))
	}
	return GRP(nodes...)
}

func RANGEI[T any](ts []T, cb func(i int, item T) NODE) NODE {
	var nodes []NODE
	for i, t := range ts {
		nodes = append(nodes, cb(i, t))
	}
	return GRP(nodes...)
}

func IF(cond bool, ifTrue NodeFn) NODE {
	if cond {
		return ifTrue()
	}
	return nil
}

func IFCachedNode(cond bool, ifTrue NODE) NODE {
	if cond {
		return ifTrue
	}
	return nil
}

func ERR(errs ...error) NODE {
	return DIV(
		CLS("alert alert-error"),
		TXT(errors.Join(errs...).Error()),
	)
}

func MINLEN(min int) NODE {
	return h.MinLength(strconv.Itoa(min))
}

func MAXLEN(max int) NODE {
	return h.MaxLength(strconv.Itoa(max))
}

var icons = map[string]string{}

func ICON(name string, children ...NODE) NODE {
	src, ok := icons[name]
	if !ok {
		parts := strings.Split(name, ":")
		if len(parts) != 2 {
			return TXT("unknown icon: " + name)
		}
		prefix := parts[0]
		icon := parts[1]
		src = fmt.Sprintf("https://api.iconify.design/%s/%s.svg", prefix, icon)

		icons[name] = src
	}

	return IMG(
		CLS("fill-white"),
		SRC(src),
		GRP(children...),
	)
}

func ROWS(rows int) NODE {
	return h.Rows(strconv.Itoa(rows))
}

func Render(w http.ResponseWriter, node NODE) {
	err := node.Render(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var SVG = h.SVG

func LINE(children ...NODE) NODE {
	return g.El("line", children...)
}

func X1(v float64) NODE {
	return ATTR("x1", fmt.Sprintf("%f", v))
}

func Y1(v float64) NODE {
	return ATTR("y1", fmt.Sprintf("%f", v))
}

func X2(v float64) NODE {
	return ATTR("x2", fmt.Sprintf("%f", v))
}

func Y2(v float64) NODE {
	return ATTR("y2", fmt.Sprintf("%f", v))
}

func TEXT(children ...NODE) NODE {
	return g.El("text", children...)
}

func DOMINANTBASELINE(v string) NODE {
	return ATTR("dominant-baseline", v)
}

func TEXTANCHOR(v string) NODE {
	return ATTR("text-anchor", v)
}

func CENTERSVGTEXT(children ...NODE) NODE {
	return GRP(
		DOMINANTBASELINE("center"),
		TEXTANCHOR("middle"),
	)
}

func WIDTH(v float64) NODE {
	return ATTR("width", fmt.Sprintf("%f", v))
}

func HEIGHT(v float64) NODE {
	return ATTR("height", fmt.Sprintf("%f", v))
}

func VIEWBOX(x, y, w, h int) NODE {
	return ATTR("viewBox", fmt.Sprintf("%d %d %d %d", x, y, w, h))
}

func VIEWBOXF(x, y, w, h float64) NODE {
	return VIEWBOX(
		int(math.Round(x)),
		int(math.Round(y)),
		int(math.Round(w)),
		int(math.Round(h)),
	)
}

func CIRCLE(children ...NODE) NODE {
	return g.El("circle", children...)
}

func CX(v float64) NODE {
	return ATTR("cx", fmt.Sprintf("%f", v))
}

func CY(v float64) NODE {
	return ATTR("cy", fmt.Sprintf("%f", v))
}

func R(v float64) NODE {
	return ATTR("r", fmt.Sprintf("%f", v))
}

func X(v float64) NODE {
	return ATTR("x", fmt.Sprintf("%f", v))
}

func Y(v float64) NODE {
	return ATTR("y", fmt.Sprintf("%f", v))
}

func STROKE(v string) NODE {
	return ATTR("stroke", v)
}

func STROKEWIDTH(v float64) NODE {
	return ATTR("stroke-width", fmt.Sprintf("%f", v))
}

func FILL(v string) NODE {
	return ATTR("fill", v)
}
