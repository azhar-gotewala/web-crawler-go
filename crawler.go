package main

import (
	"context"
	"sync"
)

type Crawler struct {
	maxWorkers   int
	urlQueue     chan string
	visited      map[string]bool
	visitedMutex sync.RWMutex
}

func NewCrawler() *Crawler {
	return &Crawler{
		maxWorkers:   5,
		urlQueue:     make(chan string, 100),
		visited:      make(map[string]bool),
		visitedMutex: sync.RWMutex{},
	}
}

func (c *Crawler) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < c.maxWorkers; i++ {
		wg.Add(1)
		go c.worker(ctx, &wg)
	}

	wg.Wait()
	return nil
}

func (c *Crawler) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case url := <-c.urlQueue:
			c.processURL(ctx, url)
		}
	}
}

func (c *Crawler) processURL(ctx context.Context, url string) error {
	if c.isVisited(url) {
		return nil
	}
	c.markVisited(url)
	// TODO: Implement URL fetching and processing
	return nil
}

func (c *Crawler) isVisited(url string) bool {
	c.visitedMutex.RLock()
	defer c.visitedMutex.RUnlock()
	return c.visited[url]
}

func (c *Crawler) markVisited(url string) {
	c.visitedMutex.Lock()
	defer c.visitedMutex.Unlock()
	c.visited[url] = true
}