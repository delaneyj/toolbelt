package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/delaneyj/toolbelt"
)

func main() {
	// create ze builder
	_nats_builder := toolbelt.NewConcreteEmbeddedNATSServerBuilder()
	// configure ze builder
	_nats_builder.SetDirectory("/var/tmp/deleteme")
	_nats_builder.SetClearData(true)
	// build
	ns, err := _nats_builder.Build()
	if err != nil {
		panic(err)
	}

	// behold ze server
	ns.NatsServer.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}
