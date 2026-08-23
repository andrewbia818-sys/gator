package config

import (
	"context"
	"fmt"
	"gator/internal/database"
	"time"

	"github.com/google/uuid"
)

// HandlerFollow takes a single url, creates a new feed follow
// record for the current user.
func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("feed url is required for follow")
	}

	feedURL := cmd.Args[0]

	feed, err := s.DB.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("failed to get feed by url: %v", err)
	}

	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %v", err)
	}

	fmt.Printf("%s is now following %s\n", user.Name, feed.Name)
	return nil
}

// Create a new unfollow command that accepts a feed's URL as an
// argument and unfollows it for the current user.
func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("feed url is required for unfollow")
	}

	feedURL := cmd.Args[0]

	feed, err := s.DB.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("failed to get feed by url: %v", err)
	}

	_, err = s.DB.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID: uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to delete feed follow: %v", err)
	}

	fmt.Printf("%s has unfollowed %s\n", user.Name, feed.Name)
	return nil
}

// HandlerGetFollowing retrieves and prints the list of feeds that the current
// user is following.
func HandlerGetFollowing(s *State, cmd Command, user database.User) error {
	follows, err := s.DB.GetFeedFollowsForUser(context.Background(), uuid.NullUUID{UUID: user.ID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to get feed follows: %v", err)
	}

	for _, f := range follows {
		fmt.Printf("%s is following %s\n", f.UserName, f.FeedName)
	}

	return nil
}

// Create a middleware function. We can use this to wrap our handler functions
// that need to know the current user. The middleware will check if the current user is set in the config.
func MiddlewareLoggedIn(
	handler func(s *State, cmd Command, user database.User) error,
) func(*State, Command) error {

	return func(s *State, cmd Command) error {
		currentUser, err := s.DB.GetUserByName(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("failed to get current user: %v", err)
		}

		return handler(s, cmd, currentUser)
	}
}
