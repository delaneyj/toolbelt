package main

import (
	"log"

	"github.com/delaneyj/toolbelt/natsrpc"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	opts := protogen.Options{
		ParamFunc: func(name, value string) error {
			log.Printf("param: %s=%s", name, value)
			return nil
		},
	}
	opts.Run(func(gen *protogen.Plugin) error {
		for _, file := range gen.Files {
			if !file.Generate {
				continue
			}

			natsrpc.Generate(gen, file)
		}
		return nil
	})
}
