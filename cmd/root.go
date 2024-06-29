package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

func Start() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	rootCmd := &cobra.Command{}
	cmd := []*cobra.Command{
		{
			Use:   "serve-all",
			Short: "Run all",
			Run: func(cmd *cobra.Command, _ []string) {
				runHttpCmd(ctx)
			},
			PreRun: func(cmd *cobra.Command, _ []string) {
				go func() {
					runQueueCmd(ctx)
				}()
				go func() {
					runCronCmd(ctx)
				}()
			},
		},
		{
			Use:   "serve-http",
			Short: "Run HTTP server",
			Run: func(cmd *cobra.Command, _ []string) {
				runHttpCmd(ctx)
			},
		},
		{
			Use:   "serve-worker",
			Short: "Run worker",
			Run: func(cmd *cobra.Command, _ []string) {
				runQueueCmd(ctx)
			},
			PreRun: func(cmd *cobra.Command, _ []string) {
				go func() {
					runCronCmd(ctx)
				}()
			},
		},
	}

	rootCmd.AddCommand(cmd...)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
