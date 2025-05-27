package config

import (
	"encoding/json"
	"fmt"
	"log" 
	"os"
)


type Config struct {
	Database DBConfig  `json:"database"` 
	Server   APIConfig `json:"server"`   
}


type DBConfig struct {
	ConnectionString string `json:"connectionString"`
}


type APIConfig struct {
	Port string `json:"port"`
}


var AppConfig Config


func LoadConfig(filePath string) (*Config, error) {
	log.Printf("INFO: Loading configuration from %s", filePath)

	configFile, err := os.Open(filePath)
	if err != nil {
	
		log.Printf("WARNING: Could not open config file '%s': %v. Using default values.", filePath, err)
	
		return nil, fmt.Errorf("could not open config file '%s': %w", filePath, err)
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	var cfg Config
	if err = decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("error decoding config file '%s': %w", filePath, err)
	}

	
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080" 
		log.Println("INFO: API port not found in config, using default '8080'.")
	}

	if cfg.Database.ConnectionString == "" {
		
		log.Println("WARNING: Database connectionString not found in config. Application might not connect to DB.")
		
	}

	AppConfig = cfg 
	return &cfg, nil
}
