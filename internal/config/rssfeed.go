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

// HandlerAddFeed takes a single name and url, creates a new feed record
// and a new feed follow record for the current user.
func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	name := cmd.Args[0]
	url := cmd.Args[1]
	now := time.Now()

	feed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
		Url:       url,
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create feed: %v", err)
	}

	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %v", err)
	}

	fmt.Printf("%s added feed %s and is now following it\n", user.Name, feed.Name)
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

// Add a function scrapefeeds. It should Get the next feed to fetch from the DB.
// timestamp of last_fetched_at should be NULL or older than 1 hour.
func ScrapeFeeds(s *State) error {
	feed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get next feed to fetch: %v", err)
	}

	rssFeed, err := FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %v", err)
	}

	fmt.Printf("Fetched feed: %s\n", rssFeed.Channel.Title)
	for _, item := range rssFeed.Channel.Item {
		fmt.Printf("- %s\n", item.Title)
	}

	err = s.DB.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return fmt.Errorf("failed to mark feed as fetched: %v", err)
	}
	return nil
}
