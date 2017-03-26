package leecher

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rusenask/ddr/messaging"
	"github.com/rusenask/ddr/types"
	"github.com/rusenask/maiden"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/nats-io/nats"

	log "github.com/Sirupsen/logrus"
)

// Leecher - default leecher interface
type Leecher interface {
	Leech(ctx context.Context) error

	Cleanup() error
	Shutdown() error
}

// DefaultLeecher - DHT/Nats based leecher
type DefaultLeecher struct {
	server string // where to find seeder/nats

	messenger *messaging.Nats
	client    *docker.Client

	distributor maiden.Distributor

	leeching map[string]*types.ImageToLeech

	mu *sync.Mutex

	sub *nats.Subscription
}

// NewDefaultLeecher - create new leecher
func NewDefaultLeecher(client *docker.Client, addr string, server string) (*DefaultLeecher, error) {
	messenger := messaging.NewNats(messaging.DefaultConfig())

	if server != "" {
		messenger.Config.Host = server
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err := messenger.InitClients(ctx)
	if err != nil {
		return nil, err
	}

	// adding initial peer, we assume that this is the same node that has Nats
	netAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", server, types.DefaultDistributionDHTPort))
	if err != nil {
		log.Fatal(err)
	}

	// this node's address
	advAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, types.DefaultDistributionDHTPort))
	if err != nil {
		log.Fatal(err)
	}

	cfg := &maiden.DHTDistributorConfig{
		MMap:         true,
		Addr:         advAddr,
		Peers:        []*net.TCPAddr{netAddr},
		UploadRate:   -1,
		DownloadRate: -1,
	}

	dhtd, err := maiden.NewDHTDistributor(cfg, client)

	if err != nil {
		return nil, err
	}

	dl := &DefaultLeecher{
		messenger:   messenger,
		server:      server,
		client:      client,
		distributor: dhtd,
		mu:          &sync.Mutex{},
		leeching:    make(map[string]*types.ImageToLeech),
	}

	return dl, nil
}

// Leech - subscribe to message queue and prepare downloading
func (l *DefaultLeecher) Leech(ctx context.Context) error {
	sub, err := l.messenger.Client.Subscribe(types.ImageDeploymentTopicName, func(m *nats.Msg) {
		req, err := types.DecodeImageToLeech(m.Data)
	HAS_ERR:
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("leecher: failed to decode image request")
			return
		}
		err = l.processLeachRequest(req)
		if err != nil {
			goto HAS_ERR
		}
	})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("leecher: failed to subsribe for image deployments")
		return err
	}
	l.sub = sub

	// requesting update
	return l.requestUpdates()
}

func (l *DefaultLeecher) alreadyLeeching(req *types.ImageToLeech) (yes bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	existing, ok := l.leeching[req.Name]
	if ok {
		if existing.ID == req.ID {
			log.WithFields(log.Fields{
				"image_id":   req.ID,
				"image_name": req.Name,
			}).Info("leecher: image is already being leeched, ignoring")
			return true
		}
		// ID is different, image was probably rebuilt
	}

	return false
}

func (l *DefaultLeecher) processLeachRequest(req *types.ImageToLeech) error {

	// checking whether we already leech it
	if l.alreadyLeeching(req) {
		return nil
	}

	// cleanup
	err := l.distributor.StopSharing(req.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"image_id":   req.ID,
			"image_name": req.Name,
		}).Warn("leecher: failed to stop sharing existing image")
	}

	// starting to pull new image
	err = l.distributor.PullImage(req.Name, req.Torrent)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"image_id":   req.ID,
			"image_name": req.Name,
		}).Error("leecher: failed to start downloading torrent")
		return err
	}

	l.mu.Lock()
	l.leeching[req.Name] = req
	l.mu.Unlock()

	return nil
}

// Cleanup - performs cleanup
func (l *DefaultLeecher) Cleanup() error {
	return l.distributor.Cleanup()
}

// Shutdown - performs shutdown
func (l *DefaultLeecher) Shutdown() error {
	l.sub.Unsubscribe()
	return l.distributor.Shutdown()
}
