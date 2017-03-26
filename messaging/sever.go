package messaging

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/nats-io/gnatsd/logger"
	"github.com/nats-io/gnatsd/server"
)

// NewServer creates a new gnatsd server from the given config and listens on
// the configured port.
func NewServer(config *Config) (*server.Server, error) {

	opts := server.Options{}
	opts.Host = config.Host
	opts.Port = config.Port
	opts.ClusterPort = config.ClusterPort
	opts.RoutesStr = config.RoutesStr
	opts.Logtime = config.Logtime
	opts.Debug = config.Debug
	opts.Trace = config.Trace

	// Configure cluster opts if explicitly set via flags.
	err := configureClusterOpts(&opts)
	if err != nil {
		server.PrintAndDie(err.Error())
	}

	s := server.New(&opts)
	if s == nil {
		return nil, fmt.Errorf("No Nats server object returned")
	}

	// Configure the logger based on the flags
	configureLogger(s, &opts)

	// Run server in Go routine.
	go s.Start()

	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		addr := s.GetListenEndpoint()
		if addr == "" {
			time.Sleep(10 * time.Millisecond)
			// Retry. We might take a little while to open a connection.
			continue
		}
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			// Retry after 50ms
			time.Sleep(50 * time.Millisecond)
			continue
		}
		conn.Close()

		return s, nil
	}

	return nil, fmt.Errorf("Unable to start Nats server")
}

func configureLogger(s *server.Server, opts *server.Options) {
	var log server.Logger

	// Check to see if stderr is being redirected and if so turn off color
	colors := true
	stat, err := os.Stderr.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		colors = false
	}
	log = logger.NewStdLogger(opts.Logtime, opts.Debug, opts.Trace, colors, true)

	s.SetLogger(log, opts.Debug, opts.Trace)
}

func configureClusterOpts(opts *server.Options) error {

	// We don't need to support a different listen address for the cluster.
	opts.ClusterHost = opts.Host

	// If we have routes, fill in here.
	if opts.RoutesStr != "" {
		opts.Routes = server.RoutesFromStr(opts.RoutesStr)
	}

	return nil
}
