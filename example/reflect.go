package main

import (
	"context"
	"fmt"
	"strings"

	greflectsvc "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"go.linka.cloud/grpc-toolkit/client"
	"go.linka.cloud/grpc-toolkit/logger"
)

func readSvcs(ctx context.Context, c client.Client) (err error) {
	log := logger.C(ctx)
	rc := greflectsvc.NewServerReflectionClient(c)
	rstream, err := rc.ServerReflectionInfo(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := rstream.CloseSend(); err2 != nil && err == nil {
			err = err2
		}
	}()
	if err = rstream.Send(&greflectsvc.ServerReflectionRequest{MessageRequest: &greflectsvc.ServerReflectionRequest_ListServices{}}); err != nil {
		return err
	}
	var rres *greflectsvc.ServerReflectionResponse
	rres, err = rstream.Recv()
	if err != nil {
		return err
	}
	rlist, ok := rres.MessageResponse.(*greflectsvc.ServerReflectionResponse_ListServicesResponse)
	if !ok {
		return fmt.Errorf("unexpected reflection response type: %T", rres.MessageResponse)
	}
	for _, v := range rlist.ListServicesResponse.Service {
		if v.Name == "grpc.reflection.v1alpha.ServerReflection" || v.Name == "grpc.reflection.v1.ServerReflection" || v.Name == "grpc.health.v1.Health" {
			continue
		}
		parts := strings.Split(v.Name, ".")
		if len(parts) < 2 {
			return fmt.Errorf("malformed service name: %s", v.Name)
		}
		pkg := strings.Join(parts[:len(parts)-1], ".")
		svc := parts[len(parts)-1]
		if err = rstream.Send(&greflectsvc.ServerReflectionRequest{MessageRequest: &greflectsvc.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: v.Name,
		}}); err != nil {
			return err
		}
		rres, err = rstream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		rfile, ok := rres.MessageResponse.(*greflectsvc.ServerReflectionResponse_FileDescriptorResponse)
		if !ok {
			return fmt.Errorf("unexpected reflection response type: %T", rres.MessageResponse)
		}
		fdps := make(map[string]*descriptorpb.DescriptorProto)
		var sdp *descriptorpb.ServiceDescriptorProto
		for _, v := range rfile.FileDescriptorResponse.FileDescriptorProto {
			fdp := &descriptorpb.FileDescriptorProto{}
			if err = proto.Unmarshal(v, fdp); err != nil {
				return err
			}
			for _, s := range fdp.GetService() {
				if fdp.GetPackage() == pkg && s.GetName() == svc {
					if sdp != nil {
						log.Warnf("service already found: %s.%s", fdp.GetPackage(), s.GetName())
						continue
					}
					sdp = s
				}
			}
			for _, m := range fdp.GetMessageType() {
				fdps[fdp.GetPackage()+"."+m.GetName()] = m
			}
		}
		if sdp == nil {
			return fmt.Errorf("%s: service not found", v.Name)
		}
		for _, m := range sdp.GetMethod() {
			log.Infof("%s: %s", v.Name, m.GetName())
		}
	}
	return nil
}
