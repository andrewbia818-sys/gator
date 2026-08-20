package config

import (
	"context"
	"database/sql"
	"fmt"
	"gator/internal/database"
	"time"

	"github.com/google/uuid"
)

// Add a state struct that holds a pointer to a config.
type State struct {
	DB     *database.Queries
	Config *Config
}

// Add a command struct. A command contains a name
// and a slice of string arguments.
type Command struct {
	Name string
	Args []string
}

// Create a commands struct. This will hold all the commands
// the CLI can handle. Add a map[string]func(*state, command) error
// field to it. This will be a map of command names to their handler functions.
type Commands struct {
	Handlers map[string]func(*State, Command) error
}

// Add a run method to the commands struct. This method runs a given command
// with the provided state if it exists.
func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.Handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("command not found: %s", cmd.Name)
	}
	return handler(s, cmd)
}

// This method registers a new handler function for a command name.
func (c *Commands) Register(name string, f func(*State, Command) error) {
	if c.Handlers == nil {
		c.Handlers = make(map[string]func(*State, Command) error)
	}
	c.Handlers[name] = f
}

// Create a login handler function: func handlerLogin(s *state, cmd command) error.
// If the command's arg's slice is empty, return an error; the login handler
// expects two arguments. A command name and a username.
func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("username is required for login")
	}

	username := cmd.Args[0]

	// Check if the user exists in the database
	_, err := s.DB.GetUserByName(context.Background(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			// User does not exist → return an error that main.go will treat as exit code 1
			return fmt.Errorf("user with name '%s' does not exist", username)
		}
		return fmt.Errorf("failed to check if user exists: %v", err)
	}

	// User exists → update config
	if err := SetUser(s.Config, username); err != nil {
		return fmt.Errorf("failed to set user: %v", err)
	}

	fmt.Printf("User has been set to: %s\n", username)
	return nil
}

// Create a register handler function and register it with the commands.
func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("username is required for register")
	}

	username := cmd.Args[0]

	// Check if the user already exists
	_, err := s.DB.GetUserByName(context.Background(), username)
	if err == nil {
		return fmt.Errorf("user with name '%s' already exists", username)
	}
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check if user exists: %v", err)
	}

	// Create the new user
	_, err = s.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	// Log the user in immediately
	if err := SetUser(s.Config, username); err != nil {
		return fmt.Errorf("failed to set user in config: %v", err)
	}

	fmt.Printf("%s has been registered and is now logged in\n", username)
	return nil
}

// Create a reset handler function and register it with the commands.
// If reset fails, exit with an error code 1 and print an error message to the console.
// If reset succeeds, print a success message to the console.
func HandlerReset(s *State, cmd Command) error {
	// Call the ResetUsers method from the database package to delete all users.
	err := s.DB.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to reset users: %v", err)
	}

	fmt.Println("All users have been reset.")
	return nil
}

// Create a GetUsers handler function and register it with the commands.
// This function will retrieve all users from the database and print their
// names to the console.
// The name of the current user will be appended with (current).
func HandlerGetUsers(s *State, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get users: %v", err)
	}

	current := s.Config.CurrentUserName

	fmt.Println("Users:")
	for _, name := range users {
		if name == current {
			fmt.Printf("- %s (current)\n", name)
		} else {
			fmt.Printf("- %s\n", name)
		}
	}

	return nil
}

// Create a agg handler function . agg command.
// Update the agg command handler function signature to now take a single
// argument: time_between_reqs set to 5s. Use the time.ParseDuration
// function to parse it into a time.Duration value.
// Print a message "Collecting feeds every <time.Duration>" when it starts.
// Use a time.Ticker to run the scrapeFeeds function once every time_between_reqs.
// Use a for loop to ensure that it runs immediately  and then every time the ticker ticks:
//func HandlerAgg(s *State, cmd Command) error {
//	if len(cmd.Args) != 1 {
//		return fmt.Errorf("invalid number of arguments for agg command")
//	}
//	timeBetweenReq := cmd.Args[0]

//	duration, err := time.ParseDuration(timeBetweenReq)
//	if err != nil {
//		return fmt.Errorf("failed to parse time between requests: %v", err)
//	}
//	// NEW CODE BELOW
//	fmt.Printf("Collecting feeds every %v\n", duration)

//	ticker := time.NewTicker(duration)
//	defer ticker.Stop()

// Run immediately, then once per tick

// ...existing code...
//for {
//    feedURL := ScrapeFeeds(s)
//    if err != nil {
//        fmt.Printf("failed to scrape feeds: %v\n", err)
//        <-ticker.C
//        continue
//    }

//    rssFeed, err := FetchFeed(context.Background(), feedURL)
//    if err != nil {
//        fmt.Printf("failed to fetch and parse RSS feed: %v\n", err)
//        <-ticker.C
//        continue
//    }

//    fmt.Printf("RSS Feed:\nTitle: %s\nLink: %s\nDescription: %s\n", rssFeed.Channel.Title, rssFeed.Channel.Link, rssFeed.Channel.Description)
//    for _, item := range rssFeed.Channel.Item {
//        fmt.Printf("\nItem:\nTitle: %s\nLink: %s\nDescription: %s\nPubDate: %s\n", item.Title, item.Link, item.Description, item.PubDate)
//    }

//	   <-ticker.C
//			return nil
//		}
//	}
//
// OLD CODE ABOVE NEW BELOW
func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("invalid number of arguments for agg command")
	}

	duration, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to parse time between requests: %v", err)
	}

	fmt.Printf("Collecting feeds every %v\n", duration)

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		err := ScrapeFeeds(s)
		if err != nil {
			fmt.Printf("failed to scrape feeds: %v\n", err)
		}

		<-ticker.C
	}
}
