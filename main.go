package main

import (
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// call the ReadConfig function to read the config file

	cfg, err := config.ReadConfig()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	// Store the config as a new instance of the state struct.
	state := &config.State{
		Config: cfg,
	}

	// Create a new instance of the commands struct with an
	// initialized map of handler functions.
	commands := &config.Commands{
		Handlers: make(map[string]func(*config.State, config.Command) error),
	}

	// Register a handler function for the login command.
	commands.Register("login", config.HandlerLogin)
	// Register a halder function for the register command.
	commands.Register("register", config.HandlerRegister)
	// Register a handler function for the reset command.
	commands.Register("reset", config.HandlerReset)
	// Register a handler function for the getusers command.
	commands.Register("users", config.HandlerGetUsers)
	// Register a handler function for the agg command.
	commands.Register("agg", config.HandlerAgg)

	// Use os.Args to get the command-line arguments passed in by the user
	args := os.Args

	// args[0] = program name
	// args[1] = command
	// args[2:] = command args

	if len(args) < 2 {
		fmt.Println("Error: No command provided. Usage: gator <command> [args]")
		os.Exit(1)
	}

	cmdName := args[1]
	cmdArgs := args[2:]

	if cmdName == "login" && len(cmdArgs) < 1 {
		fmt.Println("Error: Username is required for login. Usage: gator login <username>")
		os.Exit(1)
	}
	var dbURL = "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()
	dbQueries := database.New(db)
	state.DB = dbQueries
	fmt.Printf("Database connection established with %s and %v\n", dbURL, dbQueries)

	err = commands.Run(state, config.Command{
		Name: cmdName,
		Args: cmdArgs,
	})
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	}

	// Read the full contents of the json file and print
	// the full contents of the json file.
	//updatedCfg, err := config.ReadConfig()
	//if err != nil {
	//fmt.Println("Error reading updated config:", err)
	//return
	//}
	//fmt.Println("Verified Updated User Name:", updatedCfg.CurrentUserName)

}
