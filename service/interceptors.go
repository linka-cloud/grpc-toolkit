package service

import (
	"go.linka.cloud/grpc/interceptors"
	"go.linka.cloud/grpc/interceptors/metadata"
)

func md(opts *options) interceptors.Interceptors {
	var pairs []string
	if opts.name != "" {
		pairs = append(pairs, "grpc-service-name", opts.name)
	}
	if opts.version != "" {
		pairs = append(pairs, "grpc-service-version", opts.version)
	}
	if len(pairs) != 0 {
		return metadata.NewInterceptors(pairs...)
	}
	return nil
}
