package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/delaneyj/toolbelt/embeddednats"
)

func main() {
	// create ze builder
	ctx := context.Background()
	ns, err := embeddednats.New(ctx,
		embeddednats.WithDirectory("/var/tmp/deleteme"),
		embeddednats.WithShouldClearData(true),
	)
	if err != nil {
		panic(err)
	}

	// behold ze server
	ns.NatsServer.Start()

	ns.WaitForServer()
	nc, err := ns.Client()
	if err != nil {
		panic(err)
	}
	nc.Publish("foo", []byte("hello world"))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}
