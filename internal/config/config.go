package config

import (
	"os"
)

type Config struct {
	WebPort        string
	WGPort         string
	DBPath         string
	ServerEndpoint string
}

func Load() *Config {
	webPort := os.Getenv("WEB_PORT")
	if webPort == "" {
		webPort = "8080"
	}

	wgPort := os.Getenv("WG_PORT")
	if wgPort == "" {
		wgPort = "51820"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/app/data/panel.db"
	}

	endpoint := os.Getenv("SERVER_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:51820"
	}

	return &Config{
		WebPort:        webPort,
		WGPort:         wgPort,
		DBPath:         dbPath,
		ServerEndpoint: endpoint,
	}
}
