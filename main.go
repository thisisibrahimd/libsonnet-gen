package main

import (
	"context"
	"log/slog"
	"os"

	cmd "github.com/thisisibrahimd/libsonnet-gen/cmd/libsonnet-gen"
)

func main() {
	rootCmd := cmd.NewRootCommand()

	if err := rootCmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("ran into an error", slog.Any("error", err))
		os.Exit(1)
	}
}
