// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package proxy_test

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.linka.cloud/grpc-toolkit/proxy"
	"go.linka.cloud/grpc-toolkit/service"
)

var (
	director proxy.StreamDirector
)

func ExampleNew() {
	dst, err := grpc.Dial("example.com")
	if err != nil {
		log.Fatalf("dialing example.org: %v", err)
	}
	proxy, _ := proxy.New(dst)
	_ = proxy
}

func ExampleRegisterService() {
	// A gRPC service with the proxying codec enabled.
	svc, _ := service.New()
	// Register a TestService with 4 of its methods explicitly.
	proxy.RegisterService(svc, director,
		"mwitkow.testproto.TestService",
		"PingEmpty", "Ping", "PingError", "PingList")
}

func ExampleTransparentHandler() {
	grpc.NewServer(
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))
}

// Provides a simple example of a director that shields internal services and dials a staging or production backend.
// This is a *very naive* implementation that creates a new connection on every request. Consider using pooling.
func ExampleStreamDirector() {
	director = func(ctx context.Context, fullMethodName string) (context.Context, grpc.ClientConnInterface, error) {
		// Make sure we never forward internal services.
		if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
			return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
		}
		md, ok := metadata.FromIncomingContext(ctx)
		// Copy the inbound metadata explicitly.
		outCtx := metadata.NewOutgoingContext(ctx, md.Copy())
		if ok {
			// Decide on which backend to dial
			if val, exists := md[":authority"]; exists && val[0] == "staging.api.example.com" {
				// Make sure we use DialContext so the dialing can be cancelled/time out together with the context.
				conn, err := grpc.DialContext(ctx, "api-service.staging.svc.local")
				return outCtx, conn, err
			} else if val, exists := md[":authority"]; exists && val[0] == "api.example.com" {
				conn, err := grpc.DialContext(ctx, "api-service.prod.svc.local")
				return outCtx, conn, err
			}
		}
		return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
	}
}
