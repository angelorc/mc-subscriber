package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/angelorc/mc-subscriber/config"
	"github.com/angelorc/mc-subscriber/server"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "bitsong-mc",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:
Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}

	rootCmd.AddCommand(
		startServerCmd(),
	)

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.EnableCommandSorting = false

	rootCmd := NewRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func startServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start rest-server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(args[0])
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logger, _ := zap.NewProductionConfig().Build()
			defer logger.Sync()

			s := server.NewServer(cfg.Mailchimp, logger)
			eg, _ := errgroup.WithContext(ctx)

			log.Printf("starting server %s...", cfg.Server.Address)
			eg.Go(func() error {
				if err := s.Start(cfg.Server.Address); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("error starting server: %v", err)
					return err
				}
				return nil
			})
			log.Printf("server started %s...", cfg.Server.Address)

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			logger.Info("gracefully shutting down")
			if err := s.ShutdownWithTimeout(10 * time.Second); err != nil {
				return fmt.Errorf("shutdown server: %w", err)
			}

			cancel()
			if err := eg.Wait(); !errors.Is(err, context.Canceled) {
				return err
			}

			return nil

		},
	}

	return cmd
}
