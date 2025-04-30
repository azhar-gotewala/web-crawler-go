package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type Crawler struct {
	maxWorkers   int
	urlQueue     chan string
	visited      map[string]bool
	visitedMutex sync.RWMutex
	baseURL      *url.URL
}

func NewCrawler() *Crawler {
	return &Crawler{
		maxWorkers:   5,
		urlQueue:     make(chan string, 100),
		visited:      make(map[string]bool),
		visitedMutex: sync.RWMutex{},
	}
}

func (c *Crawler) Start(ctx context.Context, seedURL string) error {

	parsedURL, err := url.Parse(seedURL)

	if err != nil {
		return fmt.Errorf("invalid seed URL: %v", err)
	}
	c.baseURL = parsedURL

	c.urlQueue <- seedURL

	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < c.maxWorkers; i++ {
		wg.Add(1)
		go c.worker(ctx, &wg)
	}

	go func() {
		wg.Wait()
		close(c.urlQueue)
	}()
	return nil
}

func (c *Crawler) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case url, ok := <-c.urlQueue:
			if !ok {
				//Channel is closed, exit
				return
			}
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

	fmt.Printf("Crawling: %s\n", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return fmt.Errorf("error creating request:%w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	//check if the content type is HTML
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil
	}

	//Extract Links
	links, err := extractLinks(resp.Body, url)
	if err != nil {
		return fmt.Errorf("error extracting links %w", err)
	}

	for _, link := range links {
		if c.isSameDomain(link) {
			c.urlQueue <- link
		}
	}

	return nil
}

func (c *Crawler) isSameDomain(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return parsedURL.Hostname() == c.baseURL.Hostname()
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

func extractLinks(body io.Reader, baseURL string) ([]string, error) {
	links := make([]string, 0)
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	tokenizer := html.NewTokenizer(body)
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return links, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						link, err := resolveURL(base, attr.Val)
						if err == nil && isValidURL(link) {
							links = append(links, link)
						}
						break
					}
				}
			}
		}
	}
}

func resolveURL(base *url.URL, ref string) (string, error) {
	refURL, err := url.Parse(ref)
	if err != nil {
		return "", err
	}
	resolvedURL := base.ResolveReference(refURL)
	return resolvedURL.String(), nil
}

func isValidURL(urlString string) bool {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return false
	}
	return parsedURL.Scheme == "http" || parsedURL.Scheme == "https"
}
