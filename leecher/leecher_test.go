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

func TestLeecherProcessRequest(t *testing.T) {
	testutil.StartTestingNats()
	distri := testutil.NewDummyDistributor()
	nats := messaging.NewNats(messaging.DefaultConfig())
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	testutil.Expect(t, err, nil)

	dl := &DefaultLeecher{
		messenger:   nats,
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

	err = dl.processLeachRequest(req)
	testutil.Expect(t, err, nil)

	// checking what our dummy distributor got
	testutil.Expect(t, distri.PulledImage, imageName)
	testutil.Expect(t, string(distri.PulledTorrent), string(torrentFile))
}

func TestLeecherMsgProcessing(t *testing.T) {
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

	imageIDA := "some_long_id_here"
	imageNameA := "some_long_name_here"
	torrentFileA := []byte("torrentfile")

	// prepare to respond to updates request
	reqA := &types.ImageToLeech{
		ID:      imageIDA,
		Torrent: torrentFileA,
		Name:    imageNameA,
	}

	batchReq := &types.ImagesToLeechBatch{
		Images: []*types.ImageToLeech{reqA},
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

	// direct message here

	imageID := "some_long_id_here"
	imageName := "some_long_name_here"
	torrentFile := []byte("torrentfile")
	req := &types.ImageToLeech{
		ID:      imageID,
		Torrent: torrentFile,
		Name:    imageName,
	}
	encoded, err := req.Encode()
	testutil.Expect(t, err, nil)

	err = dl.Leech(context.Background())
	testutil.Expect(t, err, nil)

	//  passing msg into it
	err = nts.Client.Publish(types.ImageDeploymentTopicName, encoded)
	testutil.Expect(t, err, nil)

	time.Sleep(10 * time.Millisecond)
	// // checking what our dummy distributor got
	testutil.Expect(t, distri.PulledImage, imageName)
	testutil.Expect(t, string(distri.PulledTorrent), string(torrentFile))
}
