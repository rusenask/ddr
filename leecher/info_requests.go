package leecher

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/rusenask/ddr/types"
)

func (l *DefaultLeecher) requestUpdates() error {

	msg, err := l.messenger.Client.Request(types.ImageDeploymentInfoRequestsTopicName, nil, 5*time.Second)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("leecher: failed to request current seeding list")
		return err
	}

	batch, err := types.DecodeImagesToLeechBatch(msg.Data)
	if err != nil {
		return err
	}

	return l.processBatch(batch)
}

func (l *DefaultLeecher) processBatch(batchReq *types.ImagesToLeechBatch) error {
	for _, v := range batchReq.Images {
		err := l.processLeachRequest(v)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"image_name": v.Name,
				"image_id":   v.ID,
			}).Error("leecher: failed to process image from the batch")
			continue
		}
	}
	return nil
}

func (l *DefaultLeecher) periodicUpdates() error {
	for _ = range time.Tick(5 * time.Second) {
		select {
		default:
			err := l.requestUpdates()
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("leecher: failed to request updates")
			}
		}
	}

	return nil
}
