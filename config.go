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

func ReadConfig(configPath string) ProxyConfig {
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
