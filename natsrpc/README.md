Protobuf plugin to generate NATS equivalent to gRPC services

```shell
go install github.com/delaneyj/toolbelt/natsrpc/cmd/protoc-gen-natsrpc@latest
```

inside your `buf.gen.yaml` file, add the following:

```yaml
version: v1

plugins:
  - plugin: natsrpc
    out: ./gen
    opt:
      - paths=source_relative
```

then run `buf generate` to generate the NATS files.
