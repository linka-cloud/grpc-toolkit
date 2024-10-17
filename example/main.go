package main

import (
	"context"

	"github.com/spf13/cobra"

	"go.linka.cloud/grpc-toolkit/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	f, opts := service.NewFlagSet()
	cmd := &cobra.Command{
		Use: "example",
		Run: func(cmd *cobra.Command, args []string) {
			run(cmd.Context(), opts)
		},
	}
	cmd.Flags().AddFlagSet(f)
	cmd.ExecuteContext(ctx)
}
