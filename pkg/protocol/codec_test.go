package protocol

import (
	"net"
	"strings"
	"testing"
)

func TestCodecRoundTripRequest(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		codec := NewCodec(client)
		err := codec.WriteRequest(Request{
			ProtocolVersion: Version,
			RequestID:       "test",
			Method:          MethodDaemonStatus,
		})
		if err != nil {
			t.Errorf("write request: %v", err)
		}
	}()

	request, err := NewCodec(server).ReadRequest()
	if err != nil {
		t.Fatalf("read request: %v", err)
	}
	if request.ProtocolVersion != Version {
		t.Fatalf("protocol version = %d, want %d", request.ProtocolVersion, Version)
	}
	if request.Method != MethodDaemonStatus {
		t.Fatalf("method = %q, want %q", request.Method, MethodDaemonStatus)
	}
}

func TestCodecRejectsOversizedFrames(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		_, _ = client.Write([]byte(strings.Repeat("x", MaxFrameSize+1) + "\n"))
	}()

	var request Request
	err := NewCodec(server).readJSON(&request)
	if err != ErrFrameTooLarge {
		t.Fatalf("error = %v, want %v", err, ErrFrameTooLarge)
	}
}

func TestCodecRoundTripResponse(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		codec := NewCodec(client)
		err := codec.WriteResponse(Response{
			ProtocolVersion: Version,
			RequestID:       "test",
			OK:              true,
		})
		if err != nil {
			t.Errorf("write response: %v", err)
		}
	}()

	response, err := NewCodec(server).ReadResponse()
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if !response.OK || response.RequestID != "test" {
		t.Fatalf("response = %+v", response)
	}
}

func TestMarshalUnmarshalParams(t *testing.T) {
	raw, err := MarshalParams(RouteIDParams{ID: "route_app"})
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	var params RouteIDParams
	if err := UnmarshalParams(raw, &params); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if params.ID != "route_app" {
		t.Fatalf("params = %+v", params)
	}
}
