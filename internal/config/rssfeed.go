package config

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Define a struct to represent RSSFeeds and a struct
// for an item in the RSS feed
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// FetchRSSFeed fetches and parses the RSS feed from the given URL
func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	resp, err := http.Get(feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch RSS feed: status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS feed response body: %v", err)
	}

	var rss RSSFeed
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %v", err)
	}

	return &rss, nil
}
