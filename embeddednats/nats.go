package embeddednats

import (
	"context"
	"log"
	"os"

	"github.com/cenkalti/backoff"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type options struct {
	DataDirectory    string
	ShouldClearData  bool
	NATSSeverOptions *server.Options
}

type Option func(*options)

func WithDirectory(dir string) Option {
	return func(o *options) {
		o.DataDirectory = dir
	}
}

func WithShouldClearData(shouldClearData bool) Option {
	return func(o *options) {
		o.ShouldClearData = shouldClearData
	}
}

func WithNATSServerOptions(natsServerOptions *server.Options) Option {
	return func(o *options) {
		o.NATSSeverOptions = natsServerOptions
	}
}

type Server struct {
	NatsServer *server.Server
}

// New creates a new embedded NATS server. Will automatically start the server
// and clean up when the context is cancelled.
func New(ctx context.Context, opts ...Option) (*Server, error) {
	options := &options{
		DataDirectory: "./data/nats",
	}
	for _, o := range opts {
		o(options)
	}

	if options.ShouldClearData {
		if err := os.RemoveAll(options.DataDirectory); err != nil {
			return nil, err
		}
	}

	if options.NATSSeverOptions == nil {
		options.NATSSeverOptions = &server.Options{
			JetStream: true,
			StoreDir:  options.DataDirectory,
		}
	}

	// Initialize new server with options
	ns, err := server.NewServer(options.NATSSeverOptions)
	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		ns.Shutdown()
	}()

	// Start the server via goroutine
	ns.Start()

	return &Server{
		NatsServer: ns,
	}, nil
}

func (n *Server) Close() error {
	if n.NatsServer != nil && n.NatsServer.Running() {
		n.NatsServer.Shutdown()
	}
	return nil
}

func (n *Server) WaitForServer() {
	b := backoff.NewExponentialBackOff()

	for {
		d := b.NextBackOff()
		ready := n.NatsServer.ReadyForConnections(d)
		if ready {
			break
		}

		log.Printf("NATS server not ready, waited %s, retrying...", d)
	}
}

func (n *Server) Client() (*nats.Conn, error) {
	return nats.Connect(n.NatsServer.ClientURL())
}
