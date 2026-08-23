package config

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"gator/internal/database"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
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
	Guid        string `xml:"guid"`
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

// Helper function to extract the first URL from a string.
// It uses a regular expression to find the first occurrence
// of a URL in the string and returns it. If no URL is found,
// it returns an empty string.
func urlFromDescription(desc string) string {
	re := regexp.MustCompile(`https?://[^\s"'>]+`)
	return re.FindString(desc)
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

	for _, item := range rssFeed.Channel.Item {
		url := item.Link
		if url == "" {
			url = item.Guid
		}
		if url == "" {
			fmt.Printf("Skipping item with no URL: %s\n", item.Title)
			continue
		}
		// fmt.Printf("Post created for feed '%v' '%v' '%v'\n", feed.Name, item.Title, item.Description)

		// Deduplication check
		_, err := s.DB.GetPostByURL(context.Background(), url)
		if err == nil {
			// Post already exists
			continue
		}

		// Only insert if not found
		_, err = s.DB.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			Url:         url,
			FeedID:      feed.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to create post for feed '%s': %v", feed.Name, err)
		}

		fmt.Printf("Post created for feed '%v' '%v' '%v'\n", feed.Name, item.Title, item.Description)
	}
	return nil
}

// Add a function HandlerBrowsePosts. It should take a limit with a default
// value of 2 as an argument and fetch the posts from the database for the cucrent user.
func HandlerBrowsePosts(s *State, cmd Command, user database.User) error {
	// Default limit
	limit := 2

	// Optional argument
	if len(cmd.Args) > 0 {
		n, err := strconv.Atoi(cmd.Args[0])
		if err != nil || n <= 0 {
			return fmt.Errorf("invalid limit: %s", cmd.Args[0])
		}
		limit = n
	}

	// Fetch posts for the current user
	posts, err := s.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("failed to get posts: %v", err)
	}

	fmt.Printf("Showing %d most recent posts for user '%s':\n\n", limit, user.Name)

	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("Content: %s\n\n", post.Description.String)
	}

	return nil
}
