// Code generated by protoc-gen-defaults. DO NOT EDIT.

package testservice

var TestServiceMethods = struct {
	PingEmpty  string
	Ping       string
	PingError  string
	PingList   string
	PingStream string
}{
	PingEmpty:  "/mwitkow.testproto.TestService/PingEmpty",
	Ping:       "/mwitkow.testproto.TestService/Ping",
	PingError:  "/mwitkow.testproto.TestService/PingError",
	PingList:   "/mwitkow.testproto.TestService/PingList",
	PingStream: "/mwitkow.testproto.TestService/PingStream",
}

var PingRequestFields = struct {
	Value string
}{
	Value: "value",
}

var PingResponseFields = struct {
	Value   string
	Counter string
}{
	Value:   "value",
	Counter: "counter",
}