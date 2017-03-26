package messaging

import (
	"fmt"

	"github.com/nats-io/nats"
)

// DefaultClientConfig returns a Config struct with defaults preset.
func DefaultClientConfig() *nats.Options {
	return &nats.DefaultOptions
}

// NewClient accepts a Config and logger, and returns a pointer to the new
// Nats connection and any error that occurred during creation.
func NewClient(config *Config) (*nats.Conn, error) {

	// Create the client
	client, err := nats.Connect(ConnectURL(config))
	if err != nil {
		return nil, err
	}

	return client, err
}

// ConnectURL returns the ocnnection URL.
func ConnectURL(config *Config) string {
	return fmt.Sprintf("nats://%s:%d", config.Host, config.Port)
}
