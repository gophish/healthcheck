package config

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/gophish/gophish/logger"
)

// DMARCPrefix is the DNS prefix used by mail servers to fetch DMARC records.
const DMARCPrefix = "_dmarc"

// DKIMPrefix is the DNS prefix used by mail servers to fetch DKIM records.
// Note: this is specified by Healthcheck when sending emails.
const DKIMPrefix = "_dkim"

type Conf struct {
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
		log.Errorf("error unmarshaling config: %s", err.Error())
		return err
	}

	// Choosing the migrations directory based on the database used.
	Config.MigrationsPath = Config.MigrationsPath + Config.DBName
	return nil
}
