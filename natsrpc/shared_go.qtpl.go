// Code generated by qtc from "shared_go.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line shared_go.qtpl:1
package natsrpc

//line shared_go.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line shared_go.qtpl:1
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line shared_go.qtpl:1
func streamgoSharedTypesTemplate(qw422016 *qt422016.Writer, pkg *packageTmplData) {
//line shared_go.qtpl:1
	qw422016.N().S(`
// Code generated by protoc-gen-go-natsrpc. DO NOT EDIT.

package `)
//line shared_go.qtpl:4
	qw422016.E().S(pkg.PackageName.Snake)
//line shared_go.qtpl:4
	qw422016.N().S(`

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

const NatsRpcErrorHeader = "error"

type NatsRpcOptions struct {
	Timeout time.Duration
}
type NatsRpcOption func(*NatsRpcOptions)

func WithTimeout(timeout time.Duration) NatsRpcOption {
	return func(opt *NatsRpcOptions) {
		opt.Timeout = timeout
	}
}

var DefaultNatsRpcOptions = func() *NatsRpcOptions {
	return &NatsRpcOptions{
		Timeout: 5 * time.Minute,
	}
}

func NewNatsRpcOptions(opts ...NatsRpcOption) *NatsRpcOptions {
	opt := DefaultNatsRpcOptions()
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func sendError(msg *nats.Msg, err error) {
    msg.RespondMsg(&nats.Msg{
        Header: nats.Header{
            NatsRpcErrorHeader: []string{err.Error()},
        },
    })
}

func sendSuccess(msg *nats.Msg, res proto.Message) {
    resBytes, err := proto.Marshal(res)
    if err != nil {
        sendError(msg, fmt.Errorf("failed to marshal response: %w", err))
        return
    }
    msg.Respond(resBytes)
}

func sendEOF(msg *nats.Msg) {
    msg.Respond(nil)
}
`)
//line shared_go.qtpl:61
}

//line shared_go.qtpl:61
func writegoSharedTypesTemplate(qq422016 qtio422016.Writer, pkg *packageTmplData) {
//line shared_go.qtpl:61
	qw422016 := qt422016.AcquireWriter(qq422016)
//line shared_go.qtpl:61
	streamgoSharedTypesTemplate(qw422016, pkg)
//line shared_go.qtpl:61
	qt422016.ReleaseWriter(qw422016)
//line shared_go.qtpl:61
}

//line shared_go.qtpl:61
func goSharedTypesTemplate(pkg *packageTmplData) string {
//line shared_go.qtpl:61
	qb422016 := qt422016.AcquireByteBuffer()
//line shared_go.qtpl:61
	writegoSharedTypesTemplate(qb422016, pkg)
//line shared_go.qtpl:61
	qs422016 := string(qb422016.B)
//line shared_go.qtpl:61
	qt422016.ReleaseByteBuffer(qb422016)
//line shared_go.qtpl:61
	return qs422016
//line shared_go.qtpl:61
}
