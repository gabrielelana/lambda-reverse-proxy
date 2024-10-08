package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer cancel()

	config := Config{}
	err := defaults.Set(&config)
	if err != nil {
		return fmt.Errorf("Unable initialize configuration: %w", err)
	}

	if len(os.Args) >= 2 {
		configurationFile := os.Args[1]

		data, err := os.ReadFile(configurationFile)
		if err != nil {
			return fmt.Errorf("Unable to read configuration file %s: %w", "./lrp.yaml", err)
		}

		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return fmt.Errorf("Unable to decode configuration file: %w", err)
		}
	}

	srv := NewServer(config)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: srv,
	}
	go func() {
		fmt.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 30*time.Second)
		defer cancel()
		fmt.Println("see you space cowboy...")
		// TODO: here you can shutdown other things
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down server: %s\n", err)
		}
	}()
	wg.Wait()

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
