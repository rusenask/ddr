package seeder

import (
	"github.com/rusenask/ddr/types"

	"github.com/nats-io/nats"

	log "github.com/Sirupsen/logrus"
)

func (s *DefaultSeeder) subscribeForInfoRequests() error {
	sub, err := s.messenger.Client.Subscribe(types.ImageDeploymentInfoRequestsTopicName, func(m *nats.Msg) {
		// TODO: maybe we would need to pass here some details about the node that sent request
		reply, err := s.processInfoRequest()
	HAS_ERR:
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("default seeder: failed to process info request")
			return
		}
		err = s.messenger.Client.Publish(m.Reply, reply)
		if err != nil {
			goto HAS_ERR
		}

		// success
		return
	})

	if err != nil {
		return err
	}

	s.infoSub = sub
	return nil
}

// processInfoRequest - encodes all currently seeded torrents
func (s *DefaultSeeder) processInfoRequest() (reply []byte, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := types.ImagesToLeechBatch{}

	for _, v := range s.seeding {
		batch.Images = append(batch.Images, &v)
	}

	return batch.Encode()
}
