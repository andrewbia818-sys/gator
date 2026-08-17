package config

import (
	"context"
	"encoding/xml"
	"fmt"
	"gator/internal/database"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
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

// Create a new command called addfeed. It takes two args: name: The name
// of the feed and url: The URL of the feed
// At the top of the handler, get the current user from the database
// and connect the feed to that user.
func HandlerAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	currentUser, err := s.DB.GetUserByName(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	now := time.Now()

	feed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(), // REQUIRED
		CreatedAt: now,        // REQUIRED
		UpdatedAt: now,        // REQUIRED
		Name:      name,
		Url:       url,
		UserID: uuid.NullUUID{
			UUID:  currentUser.ID,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create feed: %v", err)
	}

	fmt.Printf("%s has been added (URL: %s)\n", feed.Name, feed.Url)
	return nil
}

// Create a command called feeds. It takes no args. It fetches all feeds
// from the database and prints to the console, the name of the feed, the URL of the feed
// and uses GetUser to get the name of the user that created the feed.
func HandlerGetFeeds(s *State, cmd Command) error {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds: %v", err)
	}

	fmt.Println("Feeds:")
	for _, feed := range feeds {
		// Look up the user who added the feed
		user, err := s.DB.GetUser(context.Background(), feed.UserID.UUID)
		if err != nil {
			return fmt.Errorf("failed to get user for feed '%s': %v", feed.Name, err)
		}

		fmt.Printf(
			"- %s (URL: %s, added by: %s)\n",
			feed.Name,
			feed.Url,
			user.Name,
		)
	}

	return nil
}
