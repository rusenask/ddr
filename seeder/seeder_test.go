package seeder

import (
	"context"
	"net"

	"github.com/rusenask/ddr/reusable"

	"github.com/rusenask/function-compose/testutil"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
)

func TestSeedImage(t *testing.T) {
	// nats := messaging.NewNats(messaging.DefaultConfig())
	err := testutil.StartTestingNats()
	testutil.Expect(t, err, nil)
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:50000")
	testutil.Expect(t, err, nil)
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	testutil.Expect(t, err, nil)

	sdr, err := NewDefaultSeeder(client, addr, "")
	testutil.Expect(t, err, nil)

	imageName := "odise/busybox-python"
	imageID, err := reusable.GetImageID(imageName, client)
	testutil.Expect(t, err, nil)

	err = sdr.Seed(context.Background(), imageName, imageID)
	testutil.Expect(t, err, nil)

	// cleaning up
	err = sdr.StopSeeding(imageName)
	testutil.Expect(t, err, nil)

	err = sdr.Shutdown()
	testutil.Expect(t, err, nil)
}

func TestDoubleSeedImage(t *testing.T) {
	// nats := messaging.NewNats(messaging.DefaultConfig())
	err := testutil.StartTestingNats()
	testutil.Expect(t, err, nil)
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:50001")
	testutil.Expect(t, err, nil)
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	testutil.Expect(t, err, nil)

	sdr, err := NewDefaultSeeder(client, addr, "")
	testutil.Expect(t, err, nil)

	imageName := "odise/busybox-python"
	imageID, err := reusable.GetImageID(imageName, client)
	testutil.Expect(t, err, nil)

	err = sdr.Seed(context.Background(), imageName, imageID)
	testutil.Expect(t, err, nil)

	// checking that we have it in seeder map
	testutil.Expect(t, sdr.seeding[imageName].ID, imageID)

	// try seeding it again
	err = sdr.Seed(context.Background(), imageName, imageID)
	testutil.Expect(t, err, nil)

	// cleaning up
	err = sdr.StopSeeding(imageName)
	testutil.Expect(t, err, nil)

	err = sdr.Shutdown()
	testutil.Expect(t, err, nil)
}

func TestSeedNonExisting(t *testing.T) {
	// nats := messaging.NewNats(messaging.DefaultConfig())
	err := testutil.StartTestingNats()
	testutil.Expect(t, err, nil)
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:50002")
	testutil.Expect(t, err, nil)
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	testutil.Expect(t, err, nil)

	sdr, err := NewDefaultSeeder(client, addr, "")
	testutil.Expect(t, err, nil)

	imageName := "karolisr/imagethatdoesntexist"

	err = sdr.Seed(context.Background(), imageName, "random_id")
	testutil.Refute(t, err, nil)

	err = sdr.Cleanup()
	testutil.Expect(t, err, nil)

	err = sdr.Shutdown()
	testutil.Expect(t, err, nil)
}
