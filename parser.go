package main

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
)

type Parser struct{}

type ParseResult struct {
	Links []string
	Title string
	Text  string
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(content []byte) (*ParseResult, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	result := &ParseResult{
		Links: make([]string, 0),
	}

	// Extract title
	result.Title = doc.Find("title").Text()

	// Extract links
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			result.Links = append(result.Links, href)
		}
	})

	// Extract text content
	result.Text = doc.Find("body").Text()

	return result, nil
}