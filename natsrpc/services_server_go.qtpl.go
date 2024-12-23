// Code generated by qtc from "services_server_go.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line services_server_go.qtpl:1
package natsrpc

//line services_server_go.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line services_server_go.qtpl:1
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line services_server_go.qtpl:1
func streamgoServerTemplate(qw422016 *qt422016.Writer, pkg *packageTmplData) {
//line services_server_go.qtpl:1
	qw422016.N().S(`
// Code generated by protoc-gen-go-natsrpc. DO NOT EDIT.

package `)
//line services_server_go.qtpl:4
	qw422016.E().S(pkg.PackageName.Snake)
//line services_server_go.qtpl:4
	qw422016.N().S(`

import (
    "context"
    "errors"
    "fmt"
    "log"

    "github.com/nats-io/nats.go"
    "google.golang.org/protobuf/proto"
    "gopkg.in/typ.v4/sync2"
)

`)
//line services_server_go.qtpl:17
	for _, service := range pkg.Services {
//line services_server_go.qtpl:19
		nsp := service.Name.Pascal

//line services_server_go.qtpl:20
		qw422016.N().S(`
type `)
//line services_server_go.qtpl:22
		qw422016.E().S(nsp)
//line services_server_go.qtpl:22
		qw422016.N().S(`Service interface {
    OnClose() error

    //#region Methods!
`)
//line services_server_go.qtpl:26
		for _, method := range service.Methods {
//line services_server_go.qtpl:28
			cs := method.IsClientStreaming
			ss := method.IsServerStreaming

//line services_server_go.qtpl:31
			switch {
//line services_server_go.qtpl:32
			case !cs && !ss:
//line services_server_go.qtpl:32
				qw422016.N().S(`			`)
//line services_server_go.qtpl:33
				qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:33
				qw422016.N().S(`(ctx context.Context, req *`)
//line services_server_go.qtpl:33
				qw422016.E().S(method.InputType.Original)
//line services_server_go.qtpl:33
				qw422016.N().S(`) (res *`)
//line services_server_go.qtpl:33
				qw422016.E().S(method.OutputType.Original)
//line services_server_go.qtpl:33
				qw422016.N().S(`, err error) // Unary call for `)
//line services_server_go.qtpl:33
				qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:33
				qw422016.N().S(`
`)
//line services_server_go.qtpl:34
			case cs && !ss:
//line services_server_go.qtpl:34
				qw422016.N().S(`            `)
//line services_server_go.qtpl:35
				qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:35
				qw422016.N().S(`(ctx context.Context, reqCh <-chan *`)
//line services_server_go.qtpl:35
				qw422016.E().S(method.InputType.Original)
//line services_server_go.qtpl:35
				qw422016.N().S(`) (res *`)
//line services_server_go.qtpl:35
				qw422016.E().S(method.OutputType.Original)
//line services_server_go.qtpl:35
				qw422016.N().S(`, err error) // Client streaming call for `)
//line services_server_go.qtpl:35
				qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:35
				qw422016.N().S(`
`)
//line services_server_go.qtpl:36
			case !cs && ss:
//line services_server_go.qtpl:36
				qw422016.N().S(`			`)
//line services_server_go.qtpl:37
				qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:37
				qw422016.N().S(`(ctx context.Context, req *`)
//line services_server_go.qtpl:37
				qw422016.E().S(method.InputType.Original)
//line services_server_go.qtpl:37
				qw422016.N().S(`, resCh chan<- *`)
//line services_server_go.qtpl:37
				qw422016.E().S(method.OutputType.Original)
//line services_server_go.qtpl:37
				qw422016.N().S(`) (err  error) // Server streaming call for `)
//line services_server_go.qtpl:37
				qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:37
				qw422016.N().S(`
`)
//line services_server_go.qtpl:38
			case cs && ss:
//line services_server_go.qtpl:38
				qw422016.N().S(`			`)
//line services_server_go.qtpl:39
				qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:39
				qw422016.N().S(`(ctx context.Context, reqCh <-chan *`)
//line services_server_go.qtpl:39
				qw422016.E().S(method.InputType.Original)
//line services_server_go.qtpl:39
				qw422016.N().S(`, resCh chan<- *`)
//line services_server_go.qtpl:39
				qw422016.E().S(method.OutputType.Original)
//line services_server_go.qtpl:39
				qw422016.N().S(`, errCh chan<- error) error // Bidirectional streaming call for `)
//line services_server_go.qtpl:39
				qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:39
				qw422016.N().S(`
`)
//line services_server_go.qtpl:40
			}
//line services_server_go.qtpl:41
		}
//line services_server_go.qtpl:41
		qw422016.N().S(`    //#endregion
}

const `)
//line services_server_go.qtpl:45
		qw422016.E().S(nsp)
//line services_server_go.qtpl:45
		qw422016.N().S(`ServiceSubject = "`)
//line services_server_go.qtpl:45
		qw422016.E().S(service.Subject)
//line services_server_go.qtpl:45
		qw422016.N().S(`"

type `)
//line services_server_go.qtpl:47
		qw422016.E().S(nsp)
//line services_server_go.qtpl:47
		qw422016.N().S(`ServiceRunner struct {
    baseSubject string
    service `)
//line services_server_go.qtpl:49
		qw422016.E().S(nsp)
//line services_server_go.qtpl:49
		qw422016.N().S(`Service
    nc *nats.Conn
}

func New`)
//line services_server_go.qtpl:53
		qw422016.E().S(nsp)
//line services_server_go.qtpl:53
		qw422016.N().S(`ServiceRunnerSingleton(ctx context.Context, nc *nats.Conn, service `)
//line services_server_go.qtpl:53
		qw422016.E().S(nsp)
//line services_server_go.qtpl:53
		qw422016.N().S(`Service) (*`)
//line services_server_go.qtpl:53
		qw422016.E().S(nsp)
//line services_server_go.qtpl:53
		qw422016.N().S(`ServiceRunner, error) {
	return New`)
//line services_server_go.qtpl:54
		qw422016.E().S(nsp)
//line services_server_go.qtpl:54
		qw422016.N().S(`ServiceRunner(ctx, nc, service, 0)
}

func New`)
//line services_server_go.qtpl:57
		qw422016.E().S(nsp)
//line services_server_go.qtpl:57
		qw422016.N().S(`ServiceRunner(ctx context.Context, nc *nats.Conn, service `)
//line services_server_go.qtpl:57
		qw422016.E().S(nsp)
//line services_server_go.qtpl:57
		qw422016.N().S(`Service, instanceID int64) (*`)
//line services_server_go.qtpl:57
		qw422016.E().S(nsp)
//line services_server_go.qtpl:57
		qw422016.N().S(`ServiceRunner, error) {
	subjectSuffix := ""
	if instanceID > 0 {
		subjectSuffix = fmt.Sprintf(".%d", instanceID)
	}

	baseSubject := fmt.Sprintf("`)
//line services_server_go.qtpl:63
		qw422016.E().S(service.Subject)
//line services_server_go.qtpl:63
		qw422016.N().S(`%s", subjectSuffix)
`)
//line services_server_go.qtpl:64
		for _, method := range service.Methods {
//line services_server_go.qtpl:64
			qw422016.N().S(`       `)
//line services_server_go.qtpl:65
			qw422016.E().S(method.Name.Camel)
//line services_server_go.qtpl:65
			qw422016.N().S(`Subject := baseSubject + ".`)
//line services_server_go.qtpl:65
			qw422016.E().S(method.Name.Kebab)
//line services_server_go.qtpl:65
			qw422016.N().S(`"
`)
//line services_server_go.qtpl:66
		}
//line services_server_go.qtpl:66
		qw422016.N().S(`
    runner := &`)
//line services_server_go.qtpl:68
		qw422016.E().S(nsp)
//line services_server_go.qtpl:68
		qw422016.N().S(`ServiceRunner{
        service: service,
        nc: nc,
    }

`)
//line services_server_go.qtpl:73
		for _, method := range service.Methods {
//line services_server_go.qtpl:75
			subjectName := method.Name.Camel + "Subject"
			ss, cs := method.IsServerStreaming, method.IsClientStreaming

//line services_server_go.qtpl:79
			switch {
//line services_server_go.qtpl:80
			case !cs && !ss:
//line services_server_go.qtpl:80
				qw422016.N().S(`            `)
//line services_server_go.qtpl:81
				streamgoServerUnaryHandler(qw422016, subjectName, method)
//line services_server_go.qtpl:81
				qw422016.N().S(`
`)
//line services_server_go.qtpl:82
			case cs && !ss:
//line services_server_go.qtpl:82
				qw422016.N().S(`            `)
//line services_server_go.qtpl:83
				streamgoServerClientStreamHandler(qw422016, subjectName, method)
//line services_server_go.qtpl:83
				qw422016.N().S(`
`)
//line services_server_go.qtpl:84
			case !cs && ss:
//line services_server_go.qtpl:84
				qw422016.N().S(`            `)
//line services_server_go.qtpl:85
				streamgoServerServerStreamHandler(qw422016, subjectName, method)
//line services_server_go.qtpl:85
				qw422016.N().S(`
`)
//line services_server_go.qtpl:86
			case cs && ss:
//line services_server_go.qtpl:86
				qw422016.N().S(`            `)
//line services_server_go.qtpl:87
				streamgoServerBidiStreamHandler(qw422016, subjectName, method)
//line services_server_go.qtpl:87
				qw422016.N().S(`
`)
//line services_server_go.qtpl:88
			}
//line services_server_go.qtpl:89
		}
//line services_server_go.qtpl:89
		qw422016.N().S(`
    return runner,nil
}

func (runner *`)
//line services_server_go.qtpl:94
		qw422016.E().S(nsp)
//line services_server_go.qtpl:94
		qw422016.N().S(`ServiceRunner) Close() error {
    var errs []error
    if runner.service != nil {
        if err := runner.service.OnClose(); err != nil {
            errs = append(errs, err)
        }
    }

    if err := errors.Join(errs...); err != nil {
        return fmt.Errorf("failed to close runner: %w", err)
    }

    return nil
}

`)
//line services_server_go.qtpl:109
	}
//line services_server_go.qtpl:109
	qw422016.N().S(`
`)
//line services_server_go.qtpl:111
}

//line services_server_go.qtpl:111
func writegoServerTemplate(qq422016 qtio422016.Writer, pkg *packageTmplData) {
//line services_server_go.qtpl:111
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services_server_go.qtpl:111
	streamgoServerTemplate(qw422016, pkg)
//line services_server_go.qtpl:111
	qt422016.ReleaseWriter(qw422016)
//line services_server_go.qtpl:111
}

//line services_server_go.qtpl:111
func goServerTemplate(pkg *packageTmplData) string {
//line services_server_go.qtpl:111
	qb422016 := qt422016.AcquireByteBuffer()
//line services_server_go.qtpl:111
	writegoServerTemplate(qb422016, pkg)
//line services_server_go.qtpl:111
	qs422016 := string(qb422016.B)
//line services_server_go.qtpl:111
	qt422016.ReleaseByteBuffer(qb422016)
//line services_server_go.qtpl:111
	return qs422016
//line services_server_go.qtpl:111
}

//line services_server_go.qtpl:114
func streamgoServerUnaryHandler(qw422016 *qt422016.Writer, subjectName string, method *methodTmplData) {
//line services_server_go.qtpl:114
	qw422016.N().S(`
// Unary call for `)
//line services_server_go.qtpl:115
	qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:115
	qw422016.N().S(`
nc.Subscribe(`)
//line services_server_go.qtpl:116
	qw422016.E().S(subjectName)
//line services_server_go.qtpl:116
	qw422016.N().S(`, func(msg *nats.Msg) {
        req := &`)
//line services_server_go.qtpl:117
	qw422016.E().S(method.InputType.Original)
//line services_server_go.qtpl:117
	qw422016.N().S(`{}
		if err := proto.Unmarshal(msg.Data, req); err != nil {
			sendError(msg, fmt.Errorf("failed to unmarshal request: %w", err))
			return
		}

		res, err := runner.service.`)
//line services_server_go.qtpl:123
	qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:123
	qw422016.N().S(`(context.Background(),req)
		if err != nil {
			sendError(msg, err)
			return
		}
		sendSuccess(msg, res)
	})
`)
//line services_server_go.qtpl:130
}

//line services_server_go.qtpl:130
func writegoServerUnaryHandler(qq422016 qtio422016.Writer, subjectName string, method *methodTmplData) {
//line services_server_go.qtpl:130
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services_server_go.qtpl:130
	streamgoServerUnaryHandler(qw422016, subjectName, method)
//line services_server_go.qtpl:130
	qt422016.ReleaseWriter(qw422016)
//line services_server_go.qtpl:130
}

//line services_server_go.qtpl:130
func goServerUnaryHandler(subjectName string, method *methodTmplData) string {
//line services_server_go.qtpl:130
	qb422016 := qt422016.AcquireByteBuffer()
//line services_server_go.qtpl:130
	writegoServerUnaryHandler(qb422016, subjectName, method)
//line services_server_go.qtpl:130
	qs422016 := string(qb422016.B)
//line services_server_go.qtpl:130
	qt422016.ReleaseByteBuffer(qb422016)
//line services_server_go.qtpl:130
	return qs422016
//line services_server_go.qtpl:130
}

//line services_server_go.qtpl:132
func streamgoServerClientStreamHandler(qw422016 *qt422016.Writer, subjectName string, method *methodTmplData) {
//line services_server_go.qtpl:132
	qw422016.N().S(`
`)
//line services_server_go.qtpl:134
	reqChName := method.Name.Camel + "ClientReqChs"
	inputName := method.InputType.Original

//line services_server_go.qtpl:136
	qw422016.N().S(`
// Client streaming call for `)
//line services_server_go.qtpl:137
	qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:137
	qw422016.N().S(`
`)
//line services_server_go.qtpl:138
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:138
	qw422016.N().S(` := sync2.Map[string, chan *`)
//line services_server_go.qtpl:138
	qw422016.N().S(inputName)
//line services_server_go.qtpl:138
	qw422016.N().S(`]{}
nc.Subscribe(`)
//line services_server_go.qtpl:139
	qw422016.E().S(subjectName)
//line services_server_go.qtpl:139
	qw422016.N().S(`, func(msg *nats.Msg) {
		// Check for end of stream
		if len(msg.Data) == 0 {
			log.Printf("Got EOF")
			reqCh, ok := `)
//line services_server_go.qtpl:143
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:143
	qw422016.N().S(`.Load(msg.Reply)
			if !ok {
				sendError(msg, errors.New("no request channel found"))
				return
			}
			close(reqCh)
			`)
//line services_server_go.qtpl:149
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:149
	qw422016.N().S(`.Delete(msg.Reply)
			return
		}

		// Check for request
		req := &`)
//line services_server_go.qtpl:154
	qw422016.N().S(inputName)
//line services_server_go.qtpl:154
	qw422016.N().S(`{}
		if err := proto.Unmarshal(msg.Data, req); err != nil {
			sendError(msg, fmt.Errorf("failed to unmarshal request: %w", err))
			return
		}

		log.Printf("Got request: %v", req)

		// Check for request channel
		reqCh, ok := `)
//line services_server_go.qtpl:163
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:163
	qw422016.N().S(`.Load(msg.Reply)
		if !ok {
			reqCh = make(chan *`)
//line services_server_go.qtpl:165
	qw422016.N().S(inputName)
//line services_server_go.qtpl:165
	qw422016.N().S(`)

			`)
//line services_server_go.qtpl:167
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:167
	qw422016.N().S(`.Store(msg.Reply, reqCh)

			go func() {
				res, err := runner.service.`)
//line services_server_go.qtpl:170
	qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:170
	qw422016.N().S(`(context.Background(),reqCh)
				if err != nil {
					sendError(msg, err)
					return
				}
				sendSuccess(msg, res)
			}()
		}
		reqCh <- req
	})
`)
//line services_server_go.qtpl:180
}

//line services_server_go.qtpl:180
func writegoServerClientStreamHandler(qq422016 qtio422016.Writer, subjectName string, method *methodTmplData) {
//line services_server_go.qtpl:180
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services_server_go.qtpl:180
	streamgoServerClientStreamHandler(qw422016, subjectName, method)
//line services_server_go.qtpl:180
	qt422016.ReleaseWriter(qw422016)
//line services_server_go.qtpl:180
}

//line services_server_go.qtpl:180
func goServerClientStreamHandler(subjectName string, method *methodTmplData) string {
//line services_server_go.qtpl:180
	qb422016 := qt422016.AcquireByteBuffer()
//line services_server_go.qtpl:180
	writegoServerClientStreamHandler(qb422016, subjectName, method)
//line services_server_go.qtpl:180
	qs422016 := string(qb422016.B)
//line services_server_go.qtpl:180
	qt422016.ReleaseByteBuffer(qb422016)
//line services_server_go.qtpl:180
	return qs422016
//line services_server_go.qtpl:180
}

//line services_server_go.qtpl:182
func streamgoServerServerStreamHandler(qw422016 *qt422016.Writer, subjectName string, method *methodTmplData) {
//line services_server_go.qtpl:182
	qw422016.N().S(`
// Server streaming call for `)
//line services_server_go.qtpl:183
	qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:183
	qw422016.N().S(`
nc.Subscribe(`)
//line services_server_go.qtpl:184
	qw422016.E().S(subjectName)
//line services_server_go.qtpl:184
	qw422016.N().S(`, func(msg *nats.Msg) {
    req := &`)
//line services_server_go.qtpl:185
	qw422016.E().S(method.InputType.Original)
//line services_server_go.qtpl:185
	qw422016.N().S(`{}
    if err := proto.Unmarshal(msg.Data, req); err != nil {
        sendError(msg, fmt.Errorf("failed to unmarshal request: %w", err))
        return
    }

	go func() {
		resCh := make(chan *`)
//line services_server_go.qtpl:192
	qw422016.E().S(method.OutputType.Original)
//line services_server_go.qtpl:192
	qw422016.N().S(`)
		defer close(resCh)

		// Send responses to client
		go func () {
			defer sendEOF(msg)
			for {
				select {
				case res, ok := <-resCh:
					if !ok {
						return
					}
					sendSuccess(msg, res)
				}
			}
		}()

		// User defined handler, this will block until the context is done
		if err := runner.service.`)
//line services_server_go.qtpl:210
	qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:210
	qw422016.N().S(`(ctx, req, resCh); err != nil {
			sendError(msg, err)
		}
	}()
})
`)
//line services_server_go.qtpl:215
}

//line services_server_go.qtpl:215
func writegoServerServerStreamHandler(qq422016 qtio422016.Writer, subjectName string, method *methodTmplData) {
//line services_server_go.qtpl:215
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services_server_go.qtpl:215
	streamgoServerServerStreamHandler(qw422016, subjectName, method)
//line services_server_go.qtpl:215
	qt422016.ReleaseWriter(qw422016)
//line services_server_go.qtpl:215
}

//line services_server_go.qtpl:215
func goServerServerStreamHandler(subjectName string, method *methodTmplData) string {
//line services_server_go.qtpl:215
	qb422016 := qt422016.AcquireByteBuffer()
//line services_server_go.qtpl:215
	writegoServerServerStreamHandler(qb422016, subjectName, method)
//line services_server_go.qtpl:215
	qs422016 := string(qb422016.B)
//line services_server_go.qtpl:215
	qt422016.ReleaseByteBuffer(qb422016)
//line services_server_go.qtpl:215
	return qs422016
//line services_server_go.qtpl:215
}

//line services_server_go.qtpl:217
func streamgoServerBidiStreamHandler(qw422016 *qt422016.Writer, subjectName string, method *methodTmplData) {
//line services_server_go.qtpl:217
	qw422016.N().S(`
`)
//line services_server_go.qtpl:219
	reqChName := method.Name.Camel + "BiReqChs"
	inputName := method.InputType.Original

//line services_server_go.qtpl:221
	qw422016.N().S(`
// Bidirectional streaming call for `)
//line services_server_go.qtpl:222
	qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:222
	qw422016.N().S(`
`)
//line services_server_go.qtpl:223
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:223
	qw422016.N().S(` := sync2.Map[string, chan *`)
//line services_server_go.qtpl:223
	qw422016.N().S(inputName)
//line services_server_go.qtpl:223
	qw422016.N().S(`]{}
nc.Subscribe(`)
//line services_server_go.qtpl:224
	qw422016.E().S(subjectName)
//line services_server_go.qtpl:224
	qw422016.N().S(`, func(msg *nats.Msg) {
		// Check for end of stream
		if len(msg.Data) == 0 {
			reqCh, ok := `)
//line services_server_go.qtpl:227
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:227
	qw422016.N().S(`.Load(msg.Reply)
			if !ok {
				sendError(msg, errors.New("no request channel found"))
				return
			}
			close(reqCh)
			`)
//line services_server_go.qtpl:233
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:233
	qw422016.N().S(`.Delete(msg.Reply)
			return
		}

		// Check for request
		req := &`)
//line services_server_go.qtpl:238
	qw422016.N().S(inputName)
//line services_server_go.qtpl:238
	qw422016.N().S(`{}
		if err := proto.Unmarshal(msg.Data, req); err != nil {
			sendError(msg, fmt.Errorf("failed to unmarshal request: %w", err))
			return
		}

		// Check for request channel
		reqCh, ok := `)
//line services_server_go.qtpl:245
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:245
	qw422016.N().S(`.Load(msg.Reply)
		if !ok {
			reqCh = make(chan *`)
//line services_server_go.qtpl:247
	qw422016.N().S(inputName)
//line services_server_go.qtpl:247
	qw422016.N().S(`)
			`)
//line services_server_go.qtpl:248
	qw422016.E().S(reqChName)
//line services_server_go.qtpl:248
	qw422016.N().S(`.Store(msg.Reply, reqCh)

			go func() {
				defer sendEOF(msg)

				resCh := make(chan *`)
//line services_server_go.qtpl:253
	qw422016.E().S(method.OutputType.Original)
//line services_server_go.qtpl:253
	qw422016.N().S(`)
				errCh := make(chan error)

				go func() {
					for {
						select {
						case res, ok := <-resCh:
							if !ok {
								return
							}
							sendSuccess(msg, res)
						case err := <-errCh:
							sendError(msg, err)
							return
						}
					}
				}()
				if err := runner.service.`)
//line services_server_go.qtpl:270
	qw422016.E().S(method.Name.Pascal)
//line services_server_go.qtpl:270
	qw422016.N().S(`(context.Background(), reqCh, resCh, errCh); err != nil {
					sendError(msg, err)
					return
				}
			}()
		}
		reqCh <- req
	})
`)
//line services_server_go.qtpl:278
}

//line services_server_go.qtpl:278
func writegoServerBidiStreamHandler(qq422016 qtio422016.Writer, subjectName string, method *methodTmplData) {
//line services_server_go.qtpl:278
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services_server_go.qtpl:278
	streamgoServerBidiStreamHandler(qw422016, subjectName, method)
//line services_server_go.qtpl:278
	qt422016.ReleaseWriter(qw422016)
//line services_server_go.qtpl:278
}

//line services_server_go.qtpl:278
func goServerBidiStreamHandler(subjectName string, method *methodTmplData) string {
//line services_server_go.qtpl:278
	qb422016 := qt422016.AcquireByteBuffer()
//line services_server_go.qtpl:278
	writegoServerBidiStreamHandler(qb422016, subjectName, method)
//line services_server_go.qtpl:278
	qs422016 := string(qb422016.B)
//line services_server_go.qtpl:278
	qt422016.ReleaseByteBuffer(qb422016)
//line services_server_go.qtpl:278
	return qs422016
//line services_server_go.qtpl:278
}
