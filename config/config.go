package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	// The port on which the server will listen
	Port int `json:"port"`
	// Twilio Account SID
	TwilioAccountSID string `json:"twilio_account_sid"`
	// Twilio API Key
	TwilioApiKey string `json:"twilio_api_key"`
	// Twilio API Secret
	TwilioApiSecret string `json:"twilio_api_secret"`
}

// loadConfig loads the configuration from environment variables or command line flags and returns a config struct
// Command line flags take precedence over environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Load environment variables with defaults
	cfg.Port = 8080 // Default port
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		}
		// Ignore error, keep default if parsing fails
	}

	cfg.TwilioAccountSID = os.Getenv("TWILIO_ACCOUNT_SID")
	cfg.TwilioApiKey = os.Getenv("TWILIO_API_KEY")
	cfg.TwilioApiSecret = os.Getenv("TWILIO_API_SECRET")

	// Load command line flags (will override environment variables/defaults)
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Port on which the server will listen")
	flag.StringVar(&cfg.TwilioAccountSID, "twilio_account_sid", cfg.TwilioAccountSID, "Twilio Account SID")
	flag.StringVar(&cfg.TwilioApiKey, "twilio_api_key", cfg.TwilioApiKey, "Twilio API Key")
	flag.StringVar(&cfg.TwilioApiSecret, "twilio_api_secret", cfg.TwilioApiSecret, "Twilio API Secret")

	flag.Parse()

	return cfg, nil
}
