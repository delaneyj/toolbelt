package toolbelt

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/cenkalti/backoff"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	lop "github.com/samber/lo/parallel"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type EmbeddedNATsServer struct {
	NatsServer *server.Server
}

func NewEmbeddedNATsServer(ctx context.Context, clearData bool) (*EmbeddedNATsServer, error) {

	const dataDir = "./data/example"
	if clearData {
		if err := os.RemoveAll(dataDir); err != nil {
			return nil, err
		}
	}

	// Initialize new server with options
	ns, err := server.NewServer(&server.Options{
		JetStream: true,
		StoreDir:  dataDir,
		Websocket: server.WebsocketOpts{
			Port:  4443,
			NoTLS: true,
		},
		HTTPPort: 8882,
	})

	if err != nil {
		panic(err)
	}

	// Start the server via goroutine
	ns.Start()

	return &EmbeddedNATsServer{
		NatsServer: ns,
	}, nil
}

func (n *EmbeddedNATsServer) Close() error {
	if n.NatsServer != nil {
		n.NatsServer.Shutdown()
	}
	return nil
}

func (n *EmbeddedNATsServer) WaitForServer() {
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

func (n *EmbeddedNATsServer) Client() (*nats.Conn, error) {
	return nats.Connect(n.NatsServer.ClientURL())
}

type TypedNewFunc[T protoreflect.ProtoMessage] func() T
type TypedIdFunc[T protoreflect.ProtoMessage] func(T) string
type TypedKV[T proto.Message] struct {
	kv      nats.KeyValue
	newFn   TypedNewFunc[T]
	getIdFn TypedIdFunc[T]
}

func UpsertTypedKV[T protoreflect.ProtoMessage](js nats.JetStreamContext, cfg *nats.KeyValueConfig, newFn TypedNewFunc[T], idFn TypedIdFunc[T]) (*TypedKV[T], error) {
	if cfg == nil || cfg.Bucket == "" {
		return nil, fmt.Errorf("invalid config")
	}

	if idFn == nil {
		return nil, fmt.Errorf("invalid idFn")
	}

	kv, err := js.KeyValue(cfg.Bucket)
	if err != nil {
		if err != nats.ErrBucketNotFound {
			return nil, fmt.Errorf("failed to kv %s: %w", cfg.Bucket, err)
		}

		kv, err = js.CreateKeyValue(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create kv %s: %w", cfg.Bucket, err)
		}
	}

	tkv := &TypedKV[T]{
		kv:      kv,
		newFn:   newFn,
		getIdFn: idFn,
	}
	return tkv, nil
}

func (tkv *TypedKV[T]) Keys() ([]string, error) {
	keys, err := tkv.kv.Keys()
	if err != nil && err != nats.ErrNoKeysFound {
		return nil, err
	}
	return keys, nil
}

func (t *TypedKV[T]) Get(key string) (T, uint64, error) {
	entry, err := t.kv.Get(key)
	if err != nil {
		if err == nats.ErrKeyNotFound {
			var out T
			return out, 0, nil
		}
	}

	if entry == nil {
		var out T
		return out, 0, nil
	}

	b := entry.Value()
	if b == nil {
		var out T
		return out, 0, nil
	}

	out := t.newFn()
	if err = proto.Unmarshal(b, out); err != nil {
		return out, 0, err
	}

	return out, entry.Revision(), nil
}

func (tkv *TypedKV[T]) Load(keys ...string) (loaded []T, err error) {
	var errs []error
	loaded = lop.Map(keys, func(key string, i int) T {
		t, _, err := tkv.Get(key)
		if err != nil {
			errs = append(errs, err)
		}
		return t
	})

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return loaded, nil
}

func (tkv *TypedKV[T]) All() (out []T, err error) {
	keys, err := tkv.kv.Keys()
	if err != nil {
		return nil, err
	}

	return tkv.Load(keys...)
}

func (tkv *TypedKV[T]) Set(value T) (revision uint64, err error) {
	b, err := proto.Marshal(value)
	if err != nil {
		return 0, err
	}

	revision, err = tkv.kv.Put(tkv.getIdFn(value), b)
	return
}

func (tkv *TypedKV[T]) Batch(values ...T) (err error) {
	errs := lop.Map(values, func(value T, i int) error {
		_, err := tkv.Set(value)
		return err
	})

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (tkv *TypedKV[T]) Update(value T, last uint64) (revision uint64, err error) {
	b, err := proto.Marshal(value)
	if err != nil {
		return 0, err
	}

	key := tkv.getIdFn(value)
	revision, err = tkv.kv.Update(key, b, last)
	return
}

func (tkv *TypedKV[T]) DeleteKey(key string) (err error) {
	return tkv.kv.Delete(key)
}

func (tkv *TypedKV[T]) Delete(value T) (err error) {
	return tkv.kv.Delete(tkv.getIdFn(value))
}

func UpsertStream(js nats.JetStreamContext, cfg *nats.StreamConfig) (si *nats.StreamInfo, err error) {
	si, err = js.StreamInfo(cfg.Name)
	if err != nil {
		if err != nats.ErrStreamNotFound {
			return nil, fmt.Errorf("failed to get stream info: %w", err)
		}

		si, err = js.AddStream(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create stream: %w", err)
		}
	}

	return si, nil
}
