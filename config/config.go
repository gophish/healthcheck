package config

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/gophish/gophish/logger"
)

type Conf struct {
	RedisAddr      string `json:"redis_addr,omitempty"`
	DBName         string `json:"db_name,omitempty"`
	DBPath         string `json:"db_path,omitempty"`
	EmailHostname  string `json:"email_hostname,omitempty"`
	MigrationsPath string `json:"migrations_path,omitempty"`
}

var Config Conf

// LoadConfig loads the configuration from the specified filepath
func LoadConfig(filepath string) error {
	// Get the config file
	configFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Errorf("File error: %v\n", err)
		return err
	}
	err = json.Unmarshal(configFile, &Config)
	if err != nil {
		log.Error("error unmarshaling config: %s", err.Error())
		return err
	}

	// Choosing the migrations directory based on the database used.
	Config.MigrationsPath = Config.MigrationsPath + Config.DBName
	return nil
}
