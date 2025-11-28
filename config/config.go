package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	// The port on which the server will listen
	Port string
	// Twilio Account SID
	TwilioAccountSID string
	// Twilio API Key
	TwilioApiKey string
	// Twilio API Secret
	TwilioApiSecret string
	// Twilio Verify Service SID
	TwilioVerifyServiceSID string
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port: "8080", // Default port
	}

	// Lookup PORT and validate it
	if port, ok := os.LookupEnv("PORT"); ok {
		if portInt, err := strconv.Atoi(port); err == nil && portInt > 0 && portInt <= 65535 {
			cfg.Port = port
		} else {
			return nil, fmt.Errorf("invalid PORT value: %s", port)
		}
	}

	// Lookup TWILIO_ACCOUNT_SID
	if accountSid, ok := os.LookupEnv("TWILIO_ACCOUNT_SID"); ok {
		cfg.TwilioAccountSID = accountSid
	} else {
		return nil, fmt.Errorf("TWILIO_ACCOUNT_SID not set")
	}

	// Lookup TWILIO_API_KEY
	if apiKey, ok := os.LookupEnv("TWILIO_API_KEY"); ok {
		cfg.TwilioApiKey = apiKey
	} else {
		return nil, fmt.Errorf("TWILIO_API_KEY not set")
	}

	// Lookup TWILIO_API_SECRET
	if apiSecret, ok := os.LookupEnv("TWILIO_API_SECRET"); ok {
		cfg.TwilioApiSecret = apiSecret
	} else {
		return nil, fmt.Errorf("TWILIO_API_SECRET not set")
	}

	// Lookup TWILIO_VERIFY_SERVICE_SID
	if verifyServiceSid, ok := os.LookupEnv("TWILIO_VERIFY_SERVICE_SID"); ok {
		cfg.TwilioVerifyServiceSID = verifyServiceSid
	} else {
		return nil, fmt.Errorf("TWILIO_VERIFY_SERVICE_SID not set")
	}

	return cfg, nil
}
