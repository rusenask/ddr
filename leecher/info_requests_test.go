package leecher

import (
	"context"
	"sync"
	"time"

	"github.com/nats-io/nats"

	"github.com/rusenask/ddr/messaging"
	"github.com/rusenask/ddr/types"

	"github.com/rusenask/ddr/testutil"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
)

func TestRequestUpdates(t *testing.T) {
	testutil.StartTestingNats()
	distri := testutil.NewDummyDistributor()
	nts := messaging.NewNats(messaging.DefaultConfig())
	nts.InitClients(context.Background())
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	testutil.Expect(t, err, nil)

	dl := &DefaultLeecher{
		messenger:   nts,
		distributor: distri,
		client:      client,
		mu:          &sync.Mutex{},
		leeching:    make(map[string]*types.ImageToLeech),
	}

	imageID := "some_long_id_here"
	imageName := "some_long_name_here"
	torrentFile := []byte("torrentfile")
	req := &types.ImageToLeech{
		ID:      imageID,
		Torrent: torrentFile,
		Name:    imageName,
	}

	batchReq := &types.ImagesToLeechBatch{
		Images: []*types.ImageToLeech{req},
	}

	err = dl.processBatch(batchReq)
	testutil.Expect(t, err, nil)

	testutil.Expect(t, dl.leeching[imageName].ID, imageID)

	// checking distributor orders
	testutil.Expect(t, distri.PulledImage, imageName)
	testutil.Expect(t, string(distri.PulledTorrent), string(torrentFile))
}

// TestRequestUpdatesB  - test for a function that is being used to request
// updates from seeder on startup and later periodically
func TestRequestUpdatesB(t *testing.T) {
	testutil.StartTestingNats()
	distri := testutil.NewDummyDistributor()
	nts := messaging.NewNats(messaging.DefaultConfig())
	nts.InitClients(context.Background())
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	testutil.Expect(t, err, nil)

	dl := &DefaultLeecher{
		messenger:   nts,
		distributor: distri,
		client:      client,
		mu:          &sync.Mutex{},
		leeching:    make(map[string]*types.ImageToLeech),
	}

	imageID := "some_long_id_here"
	imageName := "some_long_name_here"
	torrentFile := []byte("torrentfile")
	req := &types.ImageToLeech{
		ID:      imageID,
		Torrent: torrentFile,
		Name:    imageName,
	}

	batchReq := &types.ImagesToLeechBatch{
		Images: []*types.ImageToLeech{req},
	}

	enc, err := batchReq.Encode()
	testutil.Expect(t, err, nil)

	// waiting for request
	sub, err := nts.Client.Subscribe(types.ImageDeploymentInfoRequestsTopicName, func(m *nats.Msg) {
		err = nts.Client.Publish(m.Reply, enc)
		testutil.Expect(t, err, nil)
	})
	testutil.Expect(t, err, nil)
	defer sub.Unsubscribe()

	// request updated
	err = dl.requestUpdates()
	testutil.Expect(t, err, nil)

	time.Sleep(10 * time.Millisecond)

	testutil.Expect(t, dl.leeching[imageName].ID, imageID)

	// checking distributor orders
	testutil.Expect(t, distri.PulledImage, imageName)
	testutil.Expect(t, string(distri.PulledTorrent), string(torrentFile))
}
