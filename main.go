package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()
	crawler := NewCrawler()
	if err := crawler.Start(ctx); err != nil {
		log.Fatalf("Crawler failed: %v", err)
	}
}