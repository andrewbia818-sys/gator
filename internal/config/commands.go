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

	// Check if the user already exists in the database.
	_, err := s.DB.GetUserByName(context.Background(), username)
	if err == nil {
		return fmt.Errorf("user with name '%s' already exists", username)
	}
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check if user exists: %v", err)
	}

	// Create a new user in the database.
	newUser, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	// Set the current user in the config to the given name
	// Exit with code 1 if a user with that name already exists.
	if err := SetUser(s.Config, username); err != nil {
		return fmt.Errorf("failed to set user in config: %v", err)
	}
	//Print a message that the user was created, and log the user's
	// data to the console for your own debugging.
	fmt.Printf("User created: %s\n", newUser.Name)
	return nil
}
