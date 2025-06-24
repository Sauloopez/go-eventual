package main

import (
	"context"
	"eventual/internal/core"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func migrate() {
	err := core.Migrate()
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	if len(os.Args) != 2 {
		log.Fatal("No command provided. Commands: migrate, start")
		return
	}

	command := os.Args[1]

	switch command {
	case "migrate":
		migrate()
	case "start":
		// context for cancel control
		ctx, cancel := context.WithCancel(context.Background())
		// inject context in eventual
		eventual, err := core.NewEventual(ctx)
		// handle initialization errors
		if err != nil {
			log.Fatalf("Failed to initialize: %v", err)
		}
		defer cancel()

		// Capture Ctrl + C and SIGTERM signals
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

		// listen for signals and cancel context
		go func() {
			<-signalChan
			log.Println("[LOG] Received shutdown signal. Cleaning up...")
			cancel()
		}()

		// start the service
		if err := eventual.Start(); err != nil {
			log.Fatalf("[ERROR] Failed to start: %v", err)
		}

		// await to the context being cancelled
		<-ctx.Done()
		log.Println("[LOG] Shutdown complete.")
	default:
		// handle unknown commands
		log.Fatalf("Unknown command: %s", command)
	}
}
