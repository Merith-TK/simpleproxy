package main

import (
	"log"
	"os"

	"github.com/yosuke-furukawa/json5/encoding/json5"
)

type ProxyConfig struct {
	Proxy []Proxy `json:"proxy"`
}

type Proxy struct {
	Local  string `json:"local,omitempty"`
	Remote string `json:"remote"`
	Type   string `json:"type,omitempty"` // "tcp", "udp", or "both"
}

// ReadConfig loads the configuration or generates a default one if missing.
func ReadConfig(configPath string) ProxyConfig {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("[INFO] Config file not found, generating default config at %s\n", configPath)
		defaultConfig := ProxyConfig{
			Proxy: []Proxy{
				{
					Local:  "127.0.0.1:8080",
					Remote: "127.0.0.1:8081",
					Type:   "both",
				},
			},
		}
		saveConfig(configPath, defaultConfig)
		return defaultConfig
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to read config file (%s): %v\n", configPath, err)
	}

	var config ProxyConfig
	if err := json5.Unmarshal(configData, &config); err != nil {
		log.Fatalf("[ERROR] Failed to parse config file: %v\n", err)
	}
	return config
}

// saveConfig writes the given configuration to a file.
func saveConfig(configPath string, config ProxyConfig) {
	configData, err := json5.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Fatalf("[ERROR] Failed to generate default config: %v\n", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		log.Fatalf("[ERROR] Failed to write default config file: %v\n", err)
	}
}
