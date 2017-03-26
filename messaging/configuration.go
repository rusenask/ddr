package messaging

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"

	"github.com/nats-io/nats"
)

const (
	// DefaultPort is the default port for client connections.
	DefaultPort = nats.DefaultPort

	// DefaultHost defaults to all interfaces.
	DefaultHost = "0.0.0.0"

	// DefaultClusterPort is the default cluster communication port.
	DefaultClusterPort = 8222
)

// Config holds the configuration for the messaging client and server.
type Config struct {
	Host        string
	Port        int
	ClusterPort int
	RoutesStr   string
	Logtime     bool
	Debug       bool
	Trace       bool
}

// DefaultConfig returns the default options for the messaging client & server.
func DefaultConfig() *Config {

	// initialRoute will only be set if INITIAL environment variable defined
	initialRoute := GetURL(os.Getenv("INITIAL"), DefaultClusterPort)

	return &Config{
		Host:        DefaultHost,
		Port:        DefaultPort,
		ClusterPort: DefaultClusterPort,
		RoutesStr:   initialRoute,
		Logtime:     true,
		Debug:       false,
		Trace:       false,
	}
}

// SetURL parses a string in the form nats://server:port and saves as options.
func (c *Config) SetURL(s string) error {

	u, err := url.Parse(s)
	if err != nil {
		return err
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}
	c.Host = host

	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	c.Port = p

	return nil
}

// GetURL returns a nats URL from a host & port.
func GetURL(host string, port int) string {

	if host == "" || port == 0 {
		return ""
	}
	return fmt.Sprintf("nats://%s:%d", host, port)
}
