package codec

import (
	"fmt"

	"google.golang.org/grpc/encoding"
	_ "google.golang.org/grpc/encoding/proto"
	"google.golang.org/protobuf/proto"
)

// Name is the name registered for the proto compressor.
const Name = "proto"

type Codec struct{}

type vtprotoMessage interface {
	MarshalVT() ([]byte, error)
	UnmarshalVT([]byte) error
}

type protoMessage interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func (Codec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case vtprotoMessage:
		return m.MarshalVT()
	case protoMessage:
		return m.Marshal()
	case proto.Message:
		return proto.Marshal(m)
	default:
		return nil, fmt.Errorf("failed to marshal, message is %T, want proto.Message", v)
	}
}

func (Codec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case vtprotoMessage:
		return m.UnmarshalVT(data)
	case protoMessage:
		return m.Unmarshal(data)
	case proto.Message:
		return proto.Unmarshal(data, m)
	default:
		return fmt.Errorf("failed to unmarshal, message is %T, want proto.Message", v)
	}
}

func (Codec) Name() string {
	return Name
}

func init() {
	encoding.RegisterCodec(Codec{})
}
