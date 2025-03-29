package main

import (
	"eventual/internal/core"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

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
		err := core.Start(eventual)
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
}
