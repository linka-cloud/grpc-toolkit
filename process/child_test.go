package process_test

import (
	"context"
	"os"
	"testing"

	"go.linka.cloud/pm/reexec"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/grpc-toolkit/process"
	"go.linka.cloud/grpc-toolkit/signals"
)

func TestChild(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := process.NewChild("test-child")
	if err != nil {
		t.Fatal(err)
	}
	c.Register()

	if reexec.Init() {
		os.Exit(0)
	}

	ctx = signals.SetupSignalHandlerWithContext(ctx)
	logger.C(ctx).Infof("starting host: %v", os.Args)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return c.Serve(ctx)
	})
	g.Go(func() error {
		conn, err := c.Dial(ctx, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		res, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			return err
		}
		logger.C(ctx).Infof("health check: %v", res)
		return c.Close()
	})
	if err := g.Wait(); err != nil {
		t.Error(err)
	}
}
