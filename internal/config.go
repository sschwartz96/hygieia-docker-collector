package internal

import (
	"encoding/json"
	"io"
)

type Config struct {
	MongoHost     string `json:"mongo_host"`
	MongoPort     int    `json:"mongo_port"`
	MongoDBName   string `json:"mongo_db_name"`
	MongoUser     string `json:"mongo_user"`
	MongoPassword string `json:"mongo_password"`
	MongoSSL      bool   `json:"mongo_ssl"`
	Cron          string `json:"cron"`
}

func ParseConfig(reader io.Reader) (*Config, error) {
	configData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	return &config, err
}
