package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// The config struct represents the content of the JSON
// files and should contain:
// "db_url": "connection_string_goes_here",
// "current_user_name": "username_goes_here"
type Config struct {
	DB_url          string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

// This function reads the JSON file found at
// ~/.gatorconfig.json and returns a Config struct.
func ReadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	configFilePath := filepath.Join(homeDir, ".gatorconfig.json")

	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	return &config, nil
}

// This method on the Config struct  writes
// the config struct to the JSON file  at ~/.gatorconfig.json
// after setting the current_user_name field.
func SetUser(config *Config, userName string) error {
	config.CurrentUserName = userName
	// write the updated config back to the file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	configFilePath := filepath.Join(homeDir, ".gatorconfig.json")

	file, err := os.Create(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		return fmt.Errorf("failed to encode config file: %v", err)
	}
	return nil
}
