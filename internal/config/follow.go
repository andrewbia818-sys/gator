package config

import (
	"context"
	"fmt"
	"gator/internal/database"
	"time"

	"github.com/google/uuid"
)

// HandlerFollow takes a single url, creates a new feed follow record for the current user.
func HandlerFollow(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("feed url is required for follow")
	}

	feedURL := cmd.Args[0]

	currentUser, err := s.DB.GetUserByName(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	feed, err := s.DB.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("failed to get feed by url: %v", err)
	}
	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    uuid.NullUUID{UUID: currentUser.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %v", err)
	}

	fmt.Printf("%s is now following %s\n", currentUser.Name, feed.Name)
	return nil
}

// HandlerGetFollowing  fetches all feed follows for the current user and print them to the console.
func HandlerGetFollowing(s *State, cmd Command) error {
	currentUser, err := s.DB.GetUserByName(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}
	follows, err := s.DB.GetFeedFollowsForUser(context.Background(), uuid.NullUUID{UUID: currentUser.ID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to get feed follows: %v", err)
	}

	for _, f := range follows {
		fmt.Printf("%s is following %s\n", f.UserName, f.FeedName)
	}

	return nil
}
