package main

import (
	"eventual/internal/core"
	"log"
	"os"
	"os/signal"
	"sync"
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
		log.Fatal("Not Command Provided. Commands: migrate, start")
		return
	}

	command := os.Args[1]

	switch command {
	case "migrate":
		migrate()
		break
	case "start":
		eventual, err := core.NewEventual()
		if err != nil {
			panic(err)
		}

		var wg sync.WaitGroup

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

		exitChan := make(chan bool)

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := eventual.Start()
			if err != nil {
				log.Fatalf("Error while running the application: %v", err)
			}
		}()

		go func() {
			<-signalChan
			log.Println("Received shutdown signal. Cleaning up...")
			exitChan <- true
		}()

		<-exitChan

		log.Println("Waiting for all goroutines to finish...")
		wg.Wait()

		log.Println("Application shut down gracefully.")
		break
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
