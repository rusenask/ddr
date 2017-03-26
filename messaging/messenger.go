package messaging

import (
	"fmt"
	"time"

	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"

	"context"

	log "github.com/Sirupsen/logrus"
)

// Nats container
type Nats struct {
	Config *Config
	Client *nats.Conn
	Server *server.Server
}

// NewNats returns a new Nats struct, without starting the server or clients.
func NewNats(config *Config) *Nats {
	return &Nats{
		Config: config,
	}
}

// StartServer - starts nats server
func (n *Nats) StartServer() error {
	natsServer, err := NewServer(n.Config)
	if err != nil {
		return err
	}
	n.Server = natsServer
	return nil
}

// InitClients - initiates binary client (for now)
func (n *Nats) InitClients(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to init nats clients")
		default:
			natsClient, err := NewClient(n.Config)
			if err != nil {
				log.WithFields(log.Fields{
					"error":  err,
					"server": n.Config.Host,
					"port":   n.Config.Port,
				}).Warn("failed to connect to Nats server, retrying...")
				time.Sleep(1 * time.Second)
				continue
			}
			n.Client = natsClient
			return nil
		}
	}
}

// Shutdown is used to shutdown the NATS connection
func (n *Nats) Shutdown() {
	if n.Client != nil {
		n.Client.Close()
	}

	if n.Server != nil {
		n.Server.Shutdown()
	}
}
