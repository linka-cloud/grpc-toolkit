// Code generated by protoc-gen-defaults. DO NOT EDIT.

package main

import (
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	_ *timestamppb.Timestamp
	_ *durationpb.Duration
	_ *wrapperspb.BoolValue
)

func (x *HelloRequest) Default() {
}

func (x *HelloReply) Default() {
}

func (x *HelloStreamRequest) Default() {
	if x.Count == 0 {
		x.Count = 10
	}
}
