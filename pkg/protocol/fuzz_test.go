package protocol

import (
	"bytes"
	"testing"
)

func FuzzReadRequest(f *testing.F) {
	f.Add([]byte(`{"protocolVersion":1,"requestId":"fuzz","method":"daemon.status"}` + "\n"))
	f.Add([]byte(`{"protocolVersion":999,"requestId":"fuzz","method":"unknown"}` + "\n"))
	f.Add([]byte(`not-json` + "\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		codec := NewCodec(bytes.NewBuffer(data))
		_, _ = codec.ReadRequest()
	})
}
