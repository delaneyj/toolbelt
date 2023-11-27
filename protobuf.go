package toolbelt

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func MustProtoMarshal(msg proto.Message) []byte {
	b, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func MustProtoUnmarshal(b []byte, msg proto.Message) {
	if err := proto.Unmarshal(b, msg); err != nil {
		panic(err)
	}
}

func MustProtoJSONMarshal(msg proto.Message) []byte {
	b, err := protojson.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func MustProtoJSONUnmarshal(b []byte, msg proto.Message) {
	if err := protojson.Unmarshal(b, msg); err != nil {
		panic(err)
	}
}
