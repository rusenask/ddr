package seeder

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/nats-io/nats"

	"github.com/rusenask/ddr/messaging"
	"github.com/rusenask/ddr/types"

	"github.com/rusenask/maiden"

	docker "github.com/fsouza/go-dockerclient"

	log "github.com/Sirupsen/logrus"
)

// Seeder is used to seed images so leechers can get it
type Seeder interface {
	Serve() error

	Seed(ctx context.Context, image, id string) error
	StopSeeding(image string) error

	Cleanup() error
	Shutdown() error
}

// DefaultSeeder - default DHT & Nats based seeder
type DefaultSeeder struct {
	server    string // nats server
	messenger *messaging.Nats

	infoSub *nats.Subscription

	client *docker.Client

	init bool
	mu   *sync.Mutex

	seeding map[string]types.ImageToLeech // map of currently seeded image names + image IDs

	distributor maiden.Distributor
}

// NewDefaultSeeder - create new default seeder
func NewDefaultSeeder(client *docker.Client, addr *net.TCPAddr, natsAddr string) (*DefaultSeeder, error) {
	messenger := messaging.NewNats(messaging.DefaultConfig())

	if natsAddr != "" {
		messenger.Config.Host = natsAddr
	}

	cfg := &maiden.DHTDistributorConfig{
		MMap:         true,
		Addr:         addr,
		UploadRate:   -1,
		DownloadRate: -1,
	}
	dhtd, err := maiden.NewDHTDistributor(cfg, client)

	if err != nil {
		return nil, err
	}

	ds := &DefaultSeeder{
		messenger:   messenger,
		server:      natsAddr,
		client:      client,
		distributor: dhtd,
		mu:          &sync.Mutex{},
		seeding:     make(map[string]types.ImageToLeech),
	}

	return ds, nil
}

// Cleanup - removes all torrent infohashes and deletes locally exported images
func (s *DefaultSeeder) Cleanup() error {
	return s.distributor.Cleanup()
}

// Shutdown - cleans up & shuts down
func (s *DefaultSeeder) Shutdown() error {
	err := s.distributor.Cleanup()
	if err != nil {
		return err
	}

	return s.distributor.Shutdown()
}

// StopSeeding - stops seeding particular image
func (s *DefaultSeeder) StopSeeding(image string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.distributor.StopSharing(image)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"image": image,
		}).Error("default seeder: failed to stop seeding image, got error")
		return err
	}

	// deleting from our internal mini registry
	delete(s.seeding, image)
	return nil
}

// Serve - initialises seeder
func (s *DefaultSeeder) Serve() error {
	return s.initialise()
}

func (s *DefaultSeeder) initialise() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.init {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := s.messenger.InitClients(ctx)
		if err != nil {
			return err
		}
		s.init = true

		// subscribing to events
		err = s.subscribeForInfoRequests()
		if err != nil {
			return err
		}
	}

	return nil
}

// Seed - image
func (s *DefaultSeeder) Seed(ctx context.Context, image, id string) error {
	err := s.initialise()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"image": image,
		}).Error("default seeder: failed to initialise")
		return err
	}

	// checking for existing
	imageReq, ok := s.seeding[image]

	if ok && id == imageReq.ID {
		// already seeding same image
		return nil
	}

	if ok && id != imageReq.ID {
		// cleanup, image name matches but ID is different
		err = s.distributor.StopSharing(image)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"image": image,
			}).Error("default seeder: failed to perform cleanup")
			return err
		}
	}

	torrent, err := s.distributor.ShareImage(image)
	if err != nil {
		return err
	}

	request := types.ImageToLeech{
		ID:      id,
		Name:    image,
		Torrent: torrent,
	}

	// adding image to our registry
	s.seeding[image] = request

	encoded, err := request.Encode()
	if err != nil {
		return err
	}

	return s.messenger.Client.Publish(types.ImageDeploymentTopicName, encoded)
}
