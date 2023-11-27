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

type MustProtobufHandler struct {
	isJSON bool
}

func NewProtobufHandler(isJSON bool) *MustProtobufHandler {
	return &MustProtobufHandler{isJSON: isJSON}
}

func (h *MustProtobufHandler) Marshal(msg proto.Message) []byte {
	if h.isJSON {
		return MustProtoJSONMarshal(msg)
	}
	return MustProtoMarshal(msg)
}

func (h *MustProtobufHandler) Unmarshal(b []byte, msg proto.Message) {
	if h.isJSON {
		MustProtoJSONUnmarshal(b, msg)
	} else {
		MustProtoUnmarshal(b, msg)
	}
}
