package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		fmt.Println("\n received shutdown signal, Gracefully shutting down")
		cancel()
	}()

	crawler := NewCrawler()
	if err := crawler.Start(ctx, "https://www.google.com/"); err != nil {
		log.Fatalf("Crawler failed: %v", err)
	}

	select {
	case <-ctx.Done():
		fmt.Println("Crawler Stopped due to context cancellation")
	case <-time.After(10 * time.Second):
		fmt.Println("Crawler stopped after timeout")
		cancel()
	}

	time.Sleep(2 * time.Second)
	fmt.Println("Crawler Shutdown complete")
}
