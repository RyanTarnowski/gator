package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = "/.gatorconfig.json"

func getConfigFilePath() (string, error) {
	configFilePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return configFilePath + configFileName, nil
}

func write(config *Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	fileData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilePath, fileData, 0644)
	return nil
}

func Read() (Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, err
	}

	var config = Config{}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func (config Config) SetUser(username string) error {
	config.CurrentUserName = username

	err := write(&config)
	if err != nil {
		return err
	}

	return nil
}
